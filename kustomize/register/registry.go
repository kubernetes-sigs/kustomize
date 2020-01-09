package register

func RegisterCoreKinds() {
	if err := NewAppCRDRegisterPlugin().Config(nil, []byte{}); err != nil {
		panic(err)
	}
}
