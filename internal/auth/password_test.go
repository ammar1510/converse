package auth

import (
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		password string
	}{
		{
			name:     "Common password",
			password: "password123",
		},
		{
			name:     "Empty password",
			password: "",
		},
		{
			name:     "Long password",
			password: "thisissuperlongpasswordwithmanycharactersandnumbers123456789!@#$%^&*()",
		},
		{
			name:     "Special characters",
			password: "p@$$w0rd!#%&*()_+",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Hash the password
			hash, err := HashPassword(tc.password)
			if err != nil {
				t.Fatalf("HashPassword returned an error: %v", err)
			}

			// Verify the hash is not empty
			if hash == "" {
				t.Fatal("HashPassword returned an empty hash")
			}

			// Verify the hash is different from the original password
			if hash == tc.password {
				t.Fatal("Hash is the same as the original password")
			}

			// Verify we can check the password against the hash
			if !CheckPasswordHash(tc.password, hash) {
				t.Fatal("CheckPasswordHash returned false for a valid password/hash pair")
			}

			// Verify a wrong password fails
			wrongPassword := tc.password + "wrong"
			if CheckPasswordHash(wrongPassword, hash) {
				t.Fatal("CheckPasswordHash returned true for an invalid password/hash pair")
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	// Create a known hash for testing
	password := "testpassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password for test: %v", err)
	}

	// Test cases
	testCases := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "Correct password",
			password: password,
			hash:     hash,
			expected: true,
		},
		{
			name:     "Incorrect password",
			password: "wrongpassword",
			hash:     hash,
			expected: false,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash,
			expected: false,
		},
		{
			name:     "Invalid hash",
			password: password,
			hash:     "invalid$hash$format",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CheckPasswordHash(tc.password, tc.hash)
			if result != tc.expected {
				t.Fatalf("CheckPasswordHash(%q, %q) = %v, want %v", 
					tc.password, tc.hash, result, tc.expected)
			}
		})
	}
} 