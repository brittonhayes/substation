package condition

import (
	"context"
	"testing"

	"github.com/brexhq/substation/config"
)

var contentTests = []struct {
	name      string
	inspector inspContent
	test      []byte
	expected  bool
}{
	// matching Gzip against valid Gzip header
	{
		"gzip",
		inspContent{
			Options: inspContentOptions{
				Type: "application/x-gzip",
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255},
		true,
	},
	// matching Gzip against invalid Gzip header (bytes swapped)
	{
		"!gzip",
		inspContent{
			Options: inspContentOptions{
				Type: "application/x-gzip",
			},
		},
		[]byte{255, 139, 8, 0, 0, 0, 0, 0, 0, 31},
		false,
	},
	// matching Gzip against invalid Gzip header (bytes swapped) with negation
	{
		"gzip",
		inspContent{
			condition: condition{
				Negate: true,
			},
			Options: inspContentOptions{
				Type: "application/x-gzip",
			},
		},
		[]byte{255, 139, 8, 0, 0, 0, 0, 0, 0, 31},
		true,
	},
	// matching Zip against valid Zip header
	{
		"zip",
		inspContent{
			Options: inspContentOptions{
				Type: "application/zip",
			},
		},
		[]byte{80, 75, 0o3, 0o4},
		true,
	},
	// matching Gzip against valid Zip header
	{
		"!zip",
		inspContent{
			Options: inspContentOptions{
				Type: "application/zip",
			},
		},
		[]byte{31, 139, 8, 0, 0, 0, 0, 0, 0, 255},
		false,
	},
	// matching Zip against invalid Zip header (bytes swapped)
	{
		"!zip",
		inspContent{
			Options: inspContentOptions{
				Type: "application/zip",
			},
		},
		[]byte{0o4, 75, 0o3, 80},
		false,
	},
}

func TestContent(t *testing.T) {
	ctx := context.TODO()
	capsule := config.NewCapsule()

	for _, test := range contentTests {
		var _ Inspector = test.inspector

		capsule.SetData(test.test)

		check, err := test.inspector.Inspect(ctx, capsule)
		if err != nil {
			t.Error(err)
		}

		if test.expected != check {
			t.Errorf("expected %v, got %v", test.expected, check)
		}
	}
}

func benchmarkContent(b *testing.B, inspector inspContent, capsule config.Capsule) {
	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		_, _ = inspector.Inspect(ctx, capsule)
	}
}

func BenchmarkContent(b *testing.B) {
	capsule := config.NewCapsule()
	for _, test := range contentTests {
		b.Run(test.name,
			func(b *testing.B) {
				capsule.SetData(test.test)
				benchmarkContent(b, test.inspector, capsule)
			},
		)
	}
}
