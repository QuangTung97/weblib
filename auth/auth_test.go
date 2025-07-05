package auth

type testUser struct {
	ID    int64
	Email string
}

type serviceTest struct {
	svc Service[testUser]
}

func newServiceTest() *serviceTest {
	s := &serviceTest{}
	return s
}
