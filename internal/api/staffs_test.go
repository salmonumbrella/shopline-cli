package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaffsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/staffs" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		resp := StaffsListResponse{
			Items: []Staff{
				{
					ID:           "staff_123",
					Email:        "admin@example.com",
					FirstName:    "John",
					LastName:     "Doe",
					AccountOwner: true,
					Permissions:  []string{"full"},
				},
				{
					ID:           "staff_456",
					Email:        "support@example.com",
					FirstName:    "Jane",
					LastName:     "Smith",
					AccountOwner: false,
					Permissions:  []string{"orders", "products"},
				},
			},
			Page:       1,
			PageSize:   20,
			TotalCount: 2,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	staffs, err := client.ListStaffs(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListStaffs failed: %v", err)
	}

	if len(staffs.Items) != 2 {
		t.Errorf("Expected 2 staffs, got %d", len(staffs.Items))
	}
	if staffs.Items[0].ID != "staff_123" {
		t.Errorf("Unexpected staff ID: %s", staffs.Items[0].ID)
	}
	if staffs.Items[0].Email != "admin@example.com" {
		t.Errorf("Unexpected email: %s", staffs.Items[0].Email)
	}
}

func TestStaffsListWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("Expected page=2, got %s", r.URL.Query().Get("page"))
		}
		if r.URL.Query().Get("page_size") != "10" {
			t.Errorf("Expected page_size=10, got %s", r.URL.Query().Get("page_size"))
		}

		resp := StaffsListResponse{
			Items:      []Staff{{ID: "staff_123", Email: "test@example.com"}},
			Page:       2,
			PageSize:   10,
			TotalCount: 1,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	opts := &StaffsListOptions{
		Page:     2,
		PageSize: 10,
	}
	staffs, err := client.ListStaffs(context.Background(), opts)
	if err != nil {
		t.Fatalf("ListStaffs failed: %v", err)
	}

	if len(staffs.Items) != 1 {
		t.Errorf("Expected 1 staff, got %d", len(staffs.Items))
	}
}

func TestStaffsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/staffs/staff_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		staff := Staff{
			ID:           "staff_123",
			Email:        "admin@example.com",
			FirstName:    "John",
			LastName:     "Doe",
			AccountOwner: true,
			Locale:       "en",
			Permissions:  []string{"full"},
		}
		_ = json.NewEncoder(w).Encode(staff)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	staff, err := client.GetStaff(context.Background(), "staff_123")
	if err != nil {
		t.Fatalf("GetStaff failed: %v", err)
	}

	if staff.ID != "staff_123" {
		t.Errorf("Unexpected staff ID: %s", staff.ID)
	}
	if staff.Email != "admin@example.com" {
		t.Errorf("Unexpected email: %s", staff.Email)
	}
	if !staff.AccountOwner {
		t.Error("Expected account owner to be true")
	}
}

func TestStaffsUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/staffs/staff_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		var req StaffUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		staff := Staff{
			ID:        "staff_123",
			Email:     "admin@example.com",
			FirstName: req.FirstName,
			LastName:  req.LastName,
		}
		_ = json.NewEncoder(w).Encode(staff)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	req := &StaffUpdateRequest{
		FirstName: "Updated",
		LastName:  "Name",
	}
	staff, err := client.UpdateStaff(context.Background(), "staff_123", req)
	if err != nil {
		t.Fatalf("UpdateStaff failed: %v", err)
	}

	if staff.FirstName != "Updated" {
		t.Errorf("Unexpected first name: %s", staff.FirstName)
	}
}

func TestStaffsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/staffs/staff_123" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.SetUseOpenAPI(false)

	err := client.DeleteStaff(context.Background(), "staff_123")
	if err != nil {
		t.Fatalf("DeleteStaff failed: %v", err)
	}
}

func TestGetStaffEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.GetStaff(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "staff id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}

func TestUpdateStaffEmptyID(t *testing.T) {
	client := NewClient("token")

	_, err := client.UpdateStaff(context.Background(), "", &StaffUpdateRequest{FirstName: "Test"})
	if err == nil {
		t.Error("Expected error for empty ID, got nil")
	}
	if err != nil && err.Error() != "staff id is required" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestDeleteStaffEmptyID(t *testing.T) {
	client := NewClient("token")

	testCases := []struct {
		name string
		id   string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.DeleteStaff(context.Background(), tc.id)
			if err == nil {
				t.Error("Expected error for empty ID, got nil")
			}
			if err != nil && err.Error() != "staff id is required" {
				t.Errorf("Unexpected error message: %s", err.Error())
			}
		})
	}
}
