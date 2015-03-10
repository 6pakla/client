package engine

import (
	"fmt"
	"github.com/keybase/client/go/libkb"
	"io"
)

type PGPSignEngine struct {
	arg *PGPSignArg
	libkb.Contextified
}

type PGPSignArg struct {
	Sink     io.WriteCloser
	Source   io.ReadCloser
	Binary   bool
	KeyQuery string
}

func (p *PGPSignEngine) GetPrereqs() EnginePrereqs {
	return EnginePrereqs{
		Session: true,
	}
}

func (p *PGPSignEngine) Name() string {
	return "PGPSignEngine"
}

func (p *PGPSignEngine) RequiredUIs() []libkb.UIKind {
	return []libkb.UIKind{
		libkb.SecretUIKind,
	}
}

func (s *PGPSignEngine) SubConsumers() []libkb.UIConsumer {
	return nil
}

func NewPGPSignEngine(arg *PGPSignArg) *PGPSignEngine {
	return &PGPSignEngine{arg: arg}
}

func (p *PGPSignEngine) Run(ctx *Context, args interface{}, reply interface{}) (err error) {
	var key libkb.GenericKey
	var pgp *libkb.PgpKeyBundle
	var ok bool
	var dumpTo io.WriteCloser
	var written int64

	defer func() {
		if dumpTo != nil {
			dumpTo.Close()
		}
		p.arg.Sink.Close()
		p.arg.Source.Close()
	}()

	ska := libkb.SecretKeyArg{
		Reason:   "command-line signature",
		PGPOnly:  true,
		KeyQuery: p.arg.KeyQuery,
		Ui:       ctx.SecretUI,
	}

	key, err = p.G().Keyrings.GetSecretKey(ska)

	if err != nil {
		return
	} else if pgp, ok = key.(*libkb.PgpKeyBundle); !ok {
		err = fmt.Errorf("Can only sign with PGP keys (for now)")
		return
	} else if key == nil {
		err = fmt.Errorf("No secret key available")
		return
	}

	dumpTo, err = libkb.AttachedSignWrapper(p.arg.Sink, *pgp, !p.arg.Binary)
	if err != nil {
		return
	}

	written, err = io.Copy(dumpTo, p.arg.Source)

	if err == nil && written == 0 {
		err = fmt.Errorf("Empty source file, nothing to sign")
	}

	return
}
