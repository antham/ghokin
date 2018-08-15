package ghokin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileManagerTransform(t *testing.T) {
	type scenario struct {
		filename string
		test     func(bytes.Buffer, error)
	}

	scenarios := []scenario{
		{
			"fixtures/file1.feature",
			func(buf bytes.Buffer, err error) {
				b, e := ioutil.ReadFile("fixtures/file1.feature")

				assert.NoError(t, e)
				assert.EqualValues(t, string(b[:len(b)-1]), buf.String())
			},
		},
		{
			"fixtures/",
			func(buf bytes.Buffer, err error) {
				assert.EqualError(t, err, "Parser errors:\nread fixtures/: is a directory")
			},
		},
	}

	for _, scenario := range scenarios {
		f := NewFileManager(2, 4, 6,
			map[string]string{
				"seq": "seq 1 3",
			},
		)

		scenario.test(f.Transform(scenario.filename))
	}
}

func TestFileManagerTransformAndReplace(t *testing.T) {
	type scenario struct {
		path       string
		extensions []string
		setup      func()
		test       func([]error)
	}

	scenarios := []scenario{
		{
			"/tmp/ghokin/file1.feature",
			[]string{"feature"},
			func() {
				content := []byte(`Feature: test
   test

Scenario:            scenario1
   Given       whatever
   Then                  whatever
"""
hello world
"""
`)

				assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0777))
				assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/file1.feature", content, 0777))
			},
			func(errs []error) {
				assert.Len(t, errs, 0)

				content := `Feature: test
      test

  Scenario: scenario1
    Given whatever
    Then whatever
      """
      hello world
      """`

				b, e := ioutil.ReadFile("/tmp/ghokin/file1.feature")

				assert.NoError(t, e)
				assert.EqualValues(t, content, string(b))
			},
		},
		{
			"/tmp/ghokin/",
			[]string{"feature"},
			func() {
				content := []byte(`Feature: test
   test

Scenario:   scenario%d
   Given           whatever
   Then      whatever
"""
hello world
"""`)

				assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin/test1", 0777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin/test2/test3", 0777))

				for i, f := range []string{
					"/tmp/ghokin/file1.feature",
					"/tmp/ghokin/file2.feature",
					"/tmp/ghokin/test1/file3.feature",
					"/tmp/ghokin/test1/file4.feature",
					"/tmp/ghokin/test2/test3/file5.feature",
					"/tmp/ghokin/test2/test3/file6.feature",
				} {
					assert.NoError(t, ioutil.WriteFile(f, []byte(fmt.Sprintf(string(content), i)), 0777))
				}
			},
			func(errs []error) {
				assert.Len(t, errs, 0)

				content := `Feature: test
      test

  Scenario: scenario%d
    Given whatever
    Then whatever
      """
      hello world
      """`

				for i, f := range []string{
					"/tmp/ghokin/file1.feature",
					"/tmp/ghokin/file2.feature",
					"/tmp/ghokin/test1/file3.feature",
					"/tmp/ghokin/test1/file4.feature",
					"/tmp/ghokin/test2/test3/file5.feature",
					"/tmp/ghokin/test2/test3/file6.feature",
				} {
					b, e := ioutil.ReadFile(f)

					assert.NoError(t, e)
					assert.EqualValues(t, fmt.Sprintf(content, i), string(b))
				}
			},
		},
		{
			"/tmp/ghokin/",
			[]string{"feature"},
			func() {
				content := []byte(`Feature: test
   test

Scenario:   scenario
   Given           whatever
   Then      whatever
"""
hello world
"""`)

				assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin/test1", 0777))

				assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/file1.feature", content, 0777))
				assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/file2.feature", append([]byte("whatever"), content...), 0777))
				assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/test1/file3.feature", content, 0777))
				assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/test1/file4.feature", content, 0777))
				assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/test1/file5.feature", append([]byte("whatever"), content...), 0777))
			},
			func(errs []error) {
				assert.Len(t, errs, 2)

				msgs := []string{
					`An error occurred with file "/tmp/ghokin/file2.feature" : Parser errors:
(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'whateverFeature: test'`,
					`An error occurred with file "/tmp/ghokin/test1/file5.feature" : Parser errors:
(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'whateverFeature: test'`,
				}

				for _, e := range errs {
					var match bool
					for _, msg := range msgs {
						if msg == e.Error() {
							match = true
						}
					}

					if !match {
						assert.Fail(t, "Must fail with 2 files when formatting folder")
					}
				}
			},
		},
		{
			"/tmp/ghokin/",
			[]string{"txt", "feat"},
			func() {
				content := []byte(`Feature: test
   test

Scenario:   scenario
   Given           whatever
   Then      whatever
"""
hello world
"""`)

				assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0777))

				assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/file1.feature", content, 0777))
				assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/file2.txt", content, 0777))
				assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/file3.feat", content, 0777))
			},
			func(errs []error) {
				assert.Len(t, errs, 0)

				contentFormatted := `Feature: test
      test

  Scenario: scenario
    Given whatever
    Then whatever
      """
      hello world
      """`

				contentUnformatted := `Feature: test
   test

Scenario:   scenario
   Given           whatever
   Then      whatever
"""
hello world
"""`

				for _, s := range []struct {
					filename string
					expected string
				}{
					{
						"/tmp/ghokin/file1.feature",
						contentUnformatted,
					},
					{
						"/tmp/ghokin/file2.txt",
						contentFormatted,
					},
					{
						"/tmp/ghokin/file3.feat",
						contentFormatted,
					},
				} {
					b, e := ioutil.ReadFile(s.filename)

					assert.NoError(t, e)
					assert.EqualValues(t, s.expected, string(b))
				}
			},
		},
		{
			"/tmp/ghokin",
			[]string{"feature"},
			func() {
				assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0777))
				assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/file1.txt", []byte("file1"), 0777))
				assert.NoError(t, ioutil.WriteFile("/tmp/ghokin/file2.txt", []byte("file2"), 0777))
			},
			func(errs []error) {
				assert.Len(t, errs, 0)
			},
		},
		{
			"fixtures/file.txt",
			[]string{"txt"},
			func() {},
			func(errs []error) {
				assert.Len(t, errs, 1)
				assert.EqualError(t, errs[0], "Parser errors:\n(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'whatever'")
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.setup()

		f := NewFileManager(2, 4, 6,
			map[string]string{
				"seq": "seq 1 3",
			},
		)

		scenario.test(f.TransformAndReplace(scenario.path, scenario.extensions))
	}
}
