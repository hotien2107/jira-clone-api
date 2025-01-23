package workspace

type serviceInterface interface{}

type service struct{}

func newService() serviceInterface {
	return &service{}
}
