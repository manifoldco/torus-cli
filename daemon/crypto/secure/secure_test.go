package secure

import (
	"testing"

	gm "github.com/onsi/gomega"
)

func TestGuard(t *testing.T) {
	t.Run("can create guard and get secret back", func(t *testing.T) {
		gm.RegisterTestingT(t)

		g := NewGuard()
		defer g.Destroy()
		gm.Expect(g).To(gm.Equal(current))

		b := []byte("hello")
		c := append([]byte{}, b...)
		s, err := g.Secret(b)

		gm.Expect(err).To(gm.BeNil())
		gm.Expect(s).ToNot(gm.BeNil())

		gm.Expect(b).To(gm.Equal([]byte{0, 0, 0, 0, 0}))
		gm.Expect(s.Buffer()).To(gm.Equal(c))
	})

	t.Run("guard acts as a singleton", func(t *testing.T) {
		gm.RegisterTestingT(t)

		g := NewGuard()
		g2 := NewGuard()
		defer g.Destroy()
		defer g2.Destroy()

		gm.Expect(*g).To(gm.Equal(*g2))
	})

	t.Run("can create secret and destroy it", func(t *testing.T) {
		gm.RegisterTestingT(t)

		g := NewGuard()
		defer g.Destroy()

		b := []byte("hello")
		s, err := g.Secret(b)
		gm.Expect(err).To(gm.BeNil())

		s.Destroy()
		v := s.Buffer()

		gm.Expect(len(v)).To(gm.Equal(0))
		gm.Expect(cap(v)).To(gm.Equal(0))
		gm.Expect(len(g.secrets)).To(gm.Equal(0))
	})

	t.Run("can create random secret and access it", func(t *testing.T) {
		gm.RegisterTestingT(t)

		g := NewGuard()
		defer g.Destroy()

		s, err := g.Random(16)
		gm.Expect(err).To(gm.BeNil())

		v := s.Buffer()
		gm.Expect(len(v)).To(gm.Equal(16))
		gm.Expect(v).ToNot(gm.Equal([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	})

	t.Run("can create many secrets and destroy guard", func(t *testing.T) {
		gm.RegisterTestingT(t)

		g := NewGuard()

		b := []byte("hi")
		b2 := []byte("hello")
		c := append([]byte{}, b...)
		c2 := append([]byte{}, b2...)

		s, err := g.Secret(b)
		gm.Expect(err).To(gm.BeNil())

		s2, err := g.Secret(b2)
		gm.Expect(err).To(gm.BeNil())

		gm.Expect(s.Buffer()).To(gm.Equal(c))
		gm.Expect(s2.Buffer()).To(gm.Equal(c2))

		g.Destroy()

		gm.Expect(len(g.secrets)).To(gm.Equal(0))

		v := s.Buffer()
		v2 := s.Buffer()

		gm.Expect(len(v)).To(gm.Equal(0))
		gm.Expect(cap(v)).To(gm.Equal(0))
		gm.Expect(len(v2)).To(gm.Equal(0))
		gm.Expect(cap(v2)).To(gm.Equal(0))

		gm.Expect(current).To(gm.BeNil())
	})
}
