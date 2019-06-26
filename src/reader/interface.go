package reader

import "strconv"

//see https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/mocking#mocking
type Reader interface {
	GzipProcessor(string, string, bool) (int, int64, error)
}

type SpyReader struct {
	Calls int
	Args  [][]string
}

type RealReader struct{}

func (s *SpyReader) Initialise() {
	s.Args = make([][]string, 10)
	for i := 0; i < 10; i++ {
		s.Args[i] = make([]string, 10)
	}
}

func (s *SpyReader) GzipProcessor(filePathIn string, filePathOut string, allowOverwrite bool) (int, int64, error) {
	s.Args[s.Calls][0] = filePathIn
	s.Args[s.Calls][1] = filePathOut
	s.Args[s.Calls][2] = strconv.FormatBool(allowOverwrite)
	s.Calls++
	return 1, 1, nil
}

func (r *RealReader) GzipProcessor(filePathIn string, filePathOut string, allowOverwrite bool) (int, int64, error) {
	return GzipProcessor(filePathIn, filePathOut, allowOverwrite)
}
