package ghokin

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileManagerTransform(t *testing.T) {
	type scenario struct {
		filename string
		test     func([]byte, error)
	}

	scenarios := []scenario{
		{
			"fixtures/file1.feature",
			func(buf []byte, err error) {
				b, e := os.ReadFile("fixtures/file1.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"fixtures/utf8-with-bom.feature",
			func(buf []byte, err error) {
				b, e := os.ReadFile("fixtures/utf8-with-bom.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"fixtures/file1-with-cr.feature",
			func(buf []byte, err error) {
				b, e := os.ReadFile("fixtures/file1-with-cr.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"fixtures/file1-with-crlf.feature",
			func(buf []byte, err error) {
				b, e := os.ReadFile("fixtures/file1-with-crlf.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"fixtures/iso-8859-1-encoding.input.feature",
			func(buf []byte, err error) {
				assert.NoError(t, err)
				b, e := os.ReadFile("fixtures/iso-8859-1-encoding.expected.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, string(b), string(buf))
			},
		},
		{
			"fixtures/",
			func(buf []byte, err error) {
				assert.EqualError(t, err, "read fixtures/: is a directory")
			},
		},
		{
			"fixtures/invalid.feature",
			func(buf []byte, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.filename, func(t *testing.T) {
			t.Parallel()
			f := NewFileManager(2,
				map[string]string{
					"seq": "seq 1 3",
				},
			)
			scenario.test(f.Transform(scenario.filename))
		})
	}
}

func TestFileManagerTransformAndReplace(t *testing.T) {
	type scenario struct {
		testName   string
		paths      []string
		extensions []string
		setup      func()
		test       func([]error)
	}

	scenarios := []scenario{
		{
			"Format files",
			[]string{"/tmp/ghokin/file1.feature", "/tmp/ghokin/file2.feature"},
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
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file1.feature", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file2.feature", content, 0o777))
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
      """
`

				b, e := os.ReadFile("/tmp/ghokin/file1.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, content, string(b))
				b, e = os.ReadFile("/tmp/ghokin/file2.feature")
				assert.NoError(t, e)
				assert.EqualValues(t, content, string(b))
			},
		},
		{
			"Format folders",
			[]string{"/tmp/ghokin-1/", "/tmp/ghokin-2/"},
			[]string{"feature"},
			func() {
				content := []byte(`Feature: test
        test

Scenario:   scenario%d
   Given           whatever
   Then      whatever
"""
hello world
"""
`)

				assert.NoError(t, os.RemoveAll("/tmp/ghokin-1"))
				assert.NoError(t, os.RemoveAll("/tmp/ghokin-2"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-1", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-1/test1", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-1/test2/test3", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-2/test5", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-2/test6/test7", 0o777))

				for i, f := range []string{
					"/tmp/ghokin-1/file1.feature",
					"/tmp/ghokin-1/file2.feature",
					"/tmp/ghokin-1/test1/file3.feature",
					"/tmp/ghokin-1/test1/file4.feature",
					"/tmp/ghokin-1/test2/test3/file5.feature",
					"/tmp/ghokin-1/test2/test3/file6.feature",
					"/tmp/ghokin-2/test5/file7.feature",
					"/tmp/ghokin-2/test6/test7/file8.feature",
				} {
					assert.NoError(t, os.WriteFile(f, []byte(fmt.Sprintf(string(content), i)), 0o777))
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
      """
`

				for i, f := range []string{
					"/tmp/ghokin-1/file1.feature",
					"/tmp/ghokin-1/file2.feature",
					"/tmp/ghokin-1/test1/file3.feature",
					"/tmp/ghokin-1/test1/file4.feature",
					"/tmp/ghokin-1/test2/test3/file5.feature",
					"/tmp/ghokin-1/test2/test3/file6.feature",
					"/tmp/ghokin-2/test5/file7.feature",
					"/tmp/ghokin-2/test6/test7/file8.feature",
				} {
					b, e := os.ReadFile(f)
					assert.NoError(t, e)
					assert.EqualValues(t, fmt.Sprintf(content, i), string(b))
				}
			},
		},
		{
			"Format a folder with parsing errors",
			[]string{"/tmp/ghokin/"},
			[]string{"feature"},
			func() {
				content := []byte(`Feature: test
      test

Scenario:   scenario
   Given           whatever
   Then      whatever
"""
hello world
"""
`)

				assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin/test1", 0o777))

				assert.NoError(t, os.WriteFile("/tmp/ghokin/file1.feature", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file2.feature", append([]byte("whatever"), content...), 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/test1/file3.feature", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/test1/file4.feature", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/test1/file5.feature", append([]byte("whatever"), content...), 0o777))
			},
			func(errs []error) {
				assert.Len(t, errs, 2)

				msgs := []string{
					`an error occurred with file "/tmp/ghokin/file2.feature" : Parser errors:
(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'whateverFeature: test'`,
					`an error occurred with file "/tmp/ghokin/test1/file5.feature" : Parser errors:
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
			"Format a folder and set various extensions for feature files",
			[]string{"/tmp/ghokin/"},
			[]string{"txt", "feat"},
			func() {
				content := []byte(`Feature: test
   test

Scenario:   scenario
   Given           whatever
   Then      whatever
"""
hello world
"""
`)

				assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0o777))

				assert.NoError(t, os.WriteFile("/tmp/ghokin/file1.feature", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file2.txt", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file3.feat", content, 0o777))
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
      """
`

				contentUnformatted := `Feature: test
   test

Scenario:   scenario
   Given           whatever
   Then      whatever
"""
hello world
"""
`

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
					b, e := os.ReadFile(s.filename)
					assert.NoError(t, e)
					assert.EqualValues(t, s.expected, string(b))
				}
			},
		},
		{
			"Format folder with no feature files",
			[]string{"/tmp/ghokin"},
			[]string{"feature"},
			func() {
				assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file1.txt", []byte("file1"), 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file2.txt", []byte("file2"), 0o777))
			},
			func(errs []error) {
				assert.Len(t, errs, 0)
			},
		},
		{
			"Format a file with different extension and an error",
			[]string{"fixtures/file.txt"},
			[]string{"txt"},
			func() {},
			func(errs []error) {
				assert.Len(t, errs, 1)
				assert.EqualError(t, errs[0], "Parser errors:\n(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'whatever'")
			},
		},
		{
			"Format an unexisting folder",
			[]string{"whatever/whatever"},
			[]string{"feature"},
			func() {},
			func(errs []error) {
				assert.Len(t, errs, 1)
				assert.EqualError(t, errs[0], "stat whatever/whatever: no such file or directory")
			},
		},
		{
			"Format an invalid file",
			[]string{"fixtures/invalid.feature"},
			[]string{"feature"},
			func() {},
			func(errs []error) {
				assert.Len(t, errs, 1)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.testName, func(t *testing.T) {
			scenario.setup()
			f := NewFileManager(2,
				map[string]string{
					"seq": "seq 1 3",
				},
			)
			scenario.test(f.TransformAndReplace(scenario.paths, scenario.extensions))
		})
	}
}

func TestFileManagerCheck(t *testing.T) {
	type scenario struct {
		testName   string
		paths      []string
		extensions []string
		setup      func()
		test       func([]error)
	}

	scenarios := []scenario{
		{
			"Check files wrongly formatted",
			[]string{"/tmp/ghokin/file1.feature", "/tmp/ghokin/file2.feature"},
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
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file1.feature", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file2.feature", content, 0o777))
			},
			func(errs []error) {
				assert.Len(t, errs, 2)
				assert.EqualError(t, errs[0], `an error occurred with file "/tmp/ghokin/file1.feature" : file is not properly formatted`)
				assert.EqualError(t, errs[1], `an error occurred with file "/tmp/ghokin/file2.feature" : file is not properly formatted`)
			},
		},
		{
			"Check files are correctly formatted",
			[]string{"/tmp/ghokin/file1.feature", "/tmp/ghokin/file2.feature"},
			[]string{"feature"},
			func() {
				content := []byte(`Feature: test

  Scenario: scenario
    Given whatever
    Then whatever
      """
      hello world
      """
`)

				assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file1.feature", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file2.feature", content, 0o777))
			},
			func(errs []error) {
				assert.Len(t, errs, 0)
			},
		},
		{
			"Check folders are wrongly formatted",
			[]string{"/tmp/ghokin-1/", "/tmp/ghokin-2/"},
			[]string{"feature"},
			func() {
				content := []byte(`Feature: test
   test

Scenario:   scenario%d
   Given           whatever
   Then      whatever
"""
hello world
"""
`)

				assert.NoError(t, os.RemoveAll("/tmp/ghokin-1"))
				assert.NoError(t, os.RemoveAll("/tmp/ghokin-2"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-1", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-2", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-1/test1", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-1/test2/test3", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-2/test5", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-2/test6/test7", 0o777))

				for i, f := range []string{
					"/tmp/ghokin-1/file1.feature",
					"/tmp/ghokin-1/file2.feature",
					"/tmp/ghokin-1/test1/file3.feature",
					"/tmp/ghokin-1/test1/file4.feature",
					"/tmp/ghokin-1/test2/test3/file5.feature",
					"/tmp/ghokin-1/test2/test3/file6.feature",
					"/tmp/ghokin-2/test5/file7.feature",
					"/tmp/ghokin-2/test6/test7/file8.feature",
				} {
					assert.NoError(t, os.WriteFile(f, []byte(fmt.Sprintf(string(content), i)), 0o777))
				}
			},
			func(errs []error) {
				assert.Len(t, errs, 8)

				errors := map[string]bool{
					`an error occurred with file "/tmp/ghokin-1/file1.feature" : file is not properly formatted`:             true,
					`an error occurred with file "/tmp/ghokin-1/file2.feature" : file is not properly formatted`:             true,
					`an error occurred with file "/tmp/ghokin-1/test1/file3.feature" : file is not properly formatted`:       true,
					`an error occurred with file "/tmp/ghokin-1/test1/file4.feature" : file is not properly formatted`:       true,
					`an error occurred with file "/tmp/ghokin-1/test2/test3/file5.feature" : file is not properly formatted`: true,
					`an error occurred with file "/tmp/ghokin-1/test2/test3/file6.feature" : file is not properly formatted`: true,
					`an error occurred with file "/tmp/ghokin-2/test5/file7.feature" : file is not properly formatted`:       true,
					`an error occurred with file "/tmp/ghokin-2/test6/test7/file8.feature" : file is not properly formatted`: true,
				}

				for _, err := range errs {
					if _, ok := errors[err.Error()]; !ok {
						assert.Fail(t, fmt.Sprintf("error %s doesn't exists", err.Error()))
					}
				}
			},
		},
		{
			"Check folders are correctly formatted",
			[]string{"/tmp/ghokin-1/", "/tmp/ghokin-2/"},
			[]string{"feature"},
			func() {
				content := []byte(`Feature: test

  Scenario: scenario%d
    Given whatever
    Then whatever
      """
      hello world
      """
`)

				assert.NoError(t, os.RemoveAll("/tmp/ghokin-1"))
				assert.NoError(t, os.RemoveAll("/tmp/ghokin-2"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-1", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-1/test1", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-1/test2/test3", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-2/test5", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin-2/test6/test7", 0o777))

				for i, f := range []string{
					"/tmp/ghokin-1/file1.feature",
					"/tmp/ghokin-1/file2.feature",
					"/tmp/ghokin-1/test1/file3.feature",
					"/tmp/ghokin-1/test1/file4.feature",
					"/tmp/ghokin-1/test2/test3/file5.feature",
					"/tmp/ghokin-1/test2/test3/file6.feature",
					"/tmp/ghokin-2/test5/file7.feature",
					"/tmp/ghokin-2/test6/test7/file8.feature",
				} {
					assert.NoError(t, os.WriteFile(f, []byte(fmt.Sprintf(string(content), i)), 0o777))
				}
			},
			func(errs []error) {
				assert.Len(t, errs, 0)
			},
		},
		{
			"Check folders with parsing errors",
			[]string{"/tmp/ghokin/"},
			[]string{"feature"},
			func() {
				content := []byte(`Feature: test

  Scenario: scenario
    Given whatever
    Then whatever
      """
      hello world
      """
`)

				assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0o777))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin/test1", 0o777))

				assert.NoError(t, os.WriteFile("/tmp/ghokin/file1.feature", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file2.feature", append([]byte("whatever"), content...), 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/test1/file3.feature", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/test1/file4.feature", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/test1/file5.feature", append([]byte("whatever"), content...), 0o777))
			},
			func(errs []error) {
				assert.Len(t, errs, 2)

				msgs := []string{
					`an error occurred with file "/tmp/ghokin/file2.feature" : Parser errors:
(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'whateverFeature: test'`,
					`an error occurred with file "/tmp/ghokin/test1/file5.feature" : Parser errors:
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
			"Check a folder and set various extensions for feature files",
			[]string{"/tmp/ghokin/"},
			[]string{"txt", "feat"},
			func() {
				content := []byte(`Feature: test
   test

Scenario:   scenario
   Given           whatever
   Then      whatever
"""
hello world
"""
`)

				assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0o777))

				assert.NoError(t, os.WriteFile("/tmp/ghokin/file1.feature", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file2.txt", content, 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file3.feat", content, 0o777))
			},
			func(errs []error) {
				assert.Len(t, errs, 2)

				errors := map[string]bool{
					`an error occurred with file "/tmp/ghokin/file2.txt" : file is not properly formatted`:  true,
					`an error occurred with file "/tmp/ghokin/file3.feat" : file is not properly formatted`: true,
				}

				for _, err := range errs {
					if _, ok := errors[err.Error()]; !ok {
						assert.Fail(t, fmt.Sprintf("error %s doesn't exists", err.Error()))
					}
				}
			},
		},
		{
			"Check folder with no feature files",
			[]string{"/tmp/ghokin"},
			[]string{"feature"},
			func() {
				assert.NoError(t, os.RemoveAll("/tmp/ghokin"))
				assert.NoError(t, os.MkdirAll("/tmp/ghokin", 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file1.txt", []byte("file1"), 0o777))
				assert.NoError(t, os.WriteFile("/tmp/ghokin/file2.txt", []byte("file2"), 0o777))
			},
			func(errs []error) {
				assert.Len(t, errs, 0)
			},
		},
		{
			"Check a file with different extension and an error",
			[]string{"fixtures/file.txt"},
			[]string{"txt"},
			func() {},
			func(errs []error) {
				assert.Len(t, errs, 1)
				assert.EqualError(t, errs[0], "Parser errors:\n(1:1): expected: #EOF, #Language, #TagLine, #FeatureLine, #Comment, #Empty, got 'whatever'")
			},
		},
		{
			"Check an unexisting folder",
			[]string{"whatever/whatever"},
			[]string{"feature"},
			func() {},
			func(errs []error) {
				assert.Len(t, errs, 1)
				assert.EqualError(t, errs[0], "stat whatever/whatever: no such file or directory")
			},
		},
		{
			"Check an invalid file",
			[]string{"fixtures/invalid.feature"},
			[]string{"feature"},
			func() {},
			func(errs []error) {
				assert.Len(t, errs, 1)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.testName, func(t *testing.T) {
			scenario.setup()

			f := NewFileManager(2,
				map[string]string{
					"seq": "seq 1 3",
				},
			)

			scenario.test(f.Check(scenario.paths, scenario.extensions))
		})
	}
}
