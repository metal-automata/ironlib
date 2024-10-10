package asrockrack

import (
	"context"

	"github.com/metal-automata/ironlib/model"
	"github.com/metal-automata/ironlib/utils"
)

func (a *asrockrack) SetBIOSConfiguration(context.Context, map[string]string) error {
	return nil
}

func (a *asrockrack) GetBIOSConfiguration(ctx context.Context) (map[string]string, error) {
	asrr := utils.NewAsrrBioscontrol(false)

	return asrr.GetBIOSConfiguration(ctx, model.FormatProductName(a.GetModel()))
}
