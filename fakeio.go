package fiplog


type fakeio int

func (f fakeio ) Write(p []byte) (n int, err error) {
	return len(p),nil
}

func (f fakeio) Close() error {
	return nil
}

/*
type std int

func (s std) Write(p []byte) (n int, err error) {
	os.Stdout.C
}*/
