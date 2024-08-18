package trace

type Random interface {
	Fill([]byte) []byte
}
