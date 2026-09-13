package cmd

import "testing"



func TestCheckProfanity(t *testing.T) {
	tests := []struct {
		name string
		body string
		expectedOut string
	}{
		{
			name: "No bad words",
			body: "I had a pleasant stroll through the forest",
			expectedOut: "I had a pleasant stroll through the forest",
		},
		{	
			name: "Case-insensitivity matching",
			body: "What a Kerfuffle this meeting turned into",
			expectedOut: "What a **** this meeting turned into",
		},
		{
			name: "Punctuation attached",
			body: "Stop that sharbert! Don't cause another fornax here",
			expectedOut: "Stop that sharbert! Don't cause another **** here",
		},
	}

	for _,tc := range tests{
		t.Run(tc.name,func(t *testing.T){
			out := checkProfanity(tc.body)

			if out != tc.expectedOut{
				t.Fatalf("checkprofanity() \nout: %s \nwant: %s",out,tc.expectedOut)
			}
		})
	}
}
