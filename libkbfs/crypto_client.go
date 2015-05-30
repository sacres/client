package libkbfs

import (
	"github.com/keybase/client/go/client"
	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/protocol/go"
	"github.com/maxtaco/go-framed-msgpack-rpc/rpc2"
)

type CryptoClient struct {
	CryptoCommon
	ctx    *libkb.GlobalContext
	client keybase1.GenericClient
}

var _ Crypto = (*CryptoClient)(nil)

func NewCryptoClient(codec Codec, ctx *libkb.GlobalContext) (*CryptoClient, error) {
	_, xp, err := ctx.GetSocket()
	if err != nil {
		return nil, err
	}

	srv := rpc2.NewServer(xp, libkb.WrapError)

	protocols := []rpc2.Protocol{
		client.NewSecretUIProtocol(),
	}

	for _, p := range protocols {
		if err := srv.Register(p); err != nil {
			if _, ok := err.(rpc2.AlreadyRegisteredError); !ok {
				return nil, err
			}
		}
	}

	client := rpc2.NewClient(xp, libkb.UnwrapError)
	return newCryptoClientWithClient(codec, ctx, client), nil
}

// For testing.
func newCryptoClientWithClient(codec Codec, ctx *libkb.GlobalContext, client keybase1.GenericClient) *CryptoClient {
	return &CryptoClient{CryptoCommon{codec}, ctx, client}
}

func (c *CryptoClient) Sign(msg []byte) (sigInfo SignatureInfo, err error) {
	defer func() {
		libkb.G.Log.Debug("Signed %d-byte message with %s: err=%v", len(msg), sigInfo, err)
	}()
	cc := keybase1.CryptoClient{Cli: c.client}
	ed25519SigInfo, err := cc.SignED25519(keybase1.SignED25519Arg{
		SessionID: 0,
		Msg:       msg,
		Reason:    "to use kbfs",
	})
	if err != nil {
		return
	}

	sigInfo = SignatureInfo{
		Version:      SigED25519,
		Signature:    ed25519SigInfo.Sig[:],
		VerifyingKey: VerifyingKey{libkb.NaclSigningKeyPublic(ed25519SigInfo.PublicKey).GetKid()},
	}
	return
}
