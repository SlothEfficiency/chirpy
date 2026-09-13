package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	type jwtTestInput struct {
		uuid        uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
	}

	type jwtTestOutput struct {
		uuid uuid.UUID
	}
	type jwtTestCase struct {
		name        string
		expectError bool
		input       jwtTestInput
		output      jwtTestOutput
	}

	names := []string{}
	expectErrors := []bool{}
	inputs := []jwtTestInput{}
	outputs := []jwtTestOutput{}

	// Define test cases here by using append
	testUUID := uuid.New()
	names = append(names, "Roundtrip")
	expectErrors = append(expectErrors, false)
	inputs = append(inputs, jwtTestInput{
		uuid:        testUUID,
		tokenSecret: "Test",
		expiresIn:   5 * time.Second,
	})
	outputs = append(outputs, jwtTestOutput{
		uuid: testUUID,
	})

	// Here are the testcases created
	jwtTestCases := []jwtTestCase{}
	for i, _ := range names {
		jwtTestCases = append(jwtTestCases, jwtTestCase{
			name:   names[i],
			input:  inputs[i],
			output: outputs[i],
		})
	}

	for _, testCase := range jwtTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.expectError == false {
				middle, err := MakeJWT(testCase.input.uuid, testCase.input.tokenSecret, testCase.input.expiresIn)
				if err != nil {
					t.Errorf("MakeJWT returned error: %v", err)
				}
				result, err := ValidateJWT(middle, testCase.input.tokenSecret)
				if err != nil || result != testCase.output.uuid {
					t.Errorf("ValidateJWT returned wrong result: %v, %v", result, err)
				}
			} else {
				middle, err1 := MakeJWT(testCase.input.uuid, testCase.input.tokenSecret, testCase.input.expiresIn)
				result, err2 := ValidateJWT(middle, testCase.input.tokenSecret)
				if err1 == nil && err2 == nil {
					t.Errorf("No error occured despite expected with result: %v", result)
				}
			}

		})
	}
}
