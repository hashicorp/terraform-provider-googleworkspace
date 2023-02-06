# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "impersonated_user" {
  description = "Google Workspace Admin User to impersonate"
}

// Google project to run from
data "google_project" "project" {
  project_id = "terraform-strategic-providers"
}

// Google service account to use
resource "google_service_account" "acctest-sa" {
  account_id   = "tf-acctest"
  display_name = "Acceptance Test SA"
  description  = "SA for Acceptance Testing"
  project      = data.google_project.project.project_id
}

// Impersonate the User with admin permissions
resource "google_service_account_iam_member" "tf-acctest-iam" {
  service_account_id = google_service_account.acctest-sa.id
  role               = "roles/iam.serviceAccountUser"
  member             = "user:${var.impersonated_user}"
}

// Google service account to use for access_token authentication
resource "google_service_account" "acctest-sa-impersonate" {
  account_id   = "tf-acctest-acctoken"
  display_name = "Acceptance Test SA for access token auth"
  description  = "SA for Acceptance Testing (for testing access_token auth)"
  project      = data.google_project.project.project_id
}

// Add permission to create tokens
resource "google_project_iam_member" "tf-acctest-iam-create-token" {
  project = data.google_project.project.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:${google_service_account.acctest-sa-impersonate.email}"
}

// Impersonate the User with admin permissions (with the sa-impersonate service account)
resource "google_service_account_iam_member" "tf-acctest-iam-sa" {
  service_account_id = google_service_account.acctest-sa-impersonate.id
  role               = "roles/iam.serviceAccountUser"
  member             = "user:${var.impersonated_user}"
}

// Add necessary roles for vault
resource "google_project_iam_member" "tf-acctest-sa-admin" {
  project = data.google_project.project.project_id
  role    = "roles/iam.serviceAccountAdmin"
  member  = "serviceAccount:${google_service_account.acctest-sa.email}"
}

resource "google_project_iam_member" "tf-acctest-sa-key-admin" {
  project = data.google_project.project.project_id
  role    = "roles/iam.serviceAccountKeyAdmin"
  member  = "serviceAccount:${google_service_account.acctest-sa.email}"
}

// generate the key to be used by vault
resource "google_service_account_key" "tf-acctest-key" {
  service_account_id = google_service_account.acctest-sa.name
}

// Add permission to create tokens to the vault-created service account
resource "google_project_iam_member" "tf-acctest-create-token" {
  project = data.google_project.project.project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:${vault_gcp_secret_roleset.roleset.service_account_email}"
}

// Enable the cloud resource manager service
resource "google_project_service" "resource-manager" {
  project = data.google_project.project.project_id
  service = "cloudresourcemanager.googleapis.com"
}

// Enable the admin sdk api service
resource "google_project_service" "admin" {
  project = data.google_project.project.project_id
  service = "admin.googleapis.com"
}

// Enable the group settings api service
resource "google_project_service" "group-settings" {
  project = data.google_project.project.project_id
  service = "groupssettings.googleapis.com"
}

// Enable the chrome policy api service
resource "google_project_service" "chrome-policy" {
  project = data.google_project.project.project_id
  service = "chromepolicy.googleapis.com"
}

// Enable the gmail api service
resource "google_project_service" "gmail" {
  project = data.google_project.project.project_id
  service = "gmail.googleapis.com"
}
