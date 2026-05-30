# 🦒 Jiraffe

**Jiraffe** is a fast, structured, and user-friendly Command Line Interface (CLI) for Atlassian Jira, written in Go. 

Designed with a robust service-based architecture, Jiraffe allows developers to seamlessly query projects, list issues, and view ticket details directly from the terminal without breaking focus.

---

## ✨ Features

* **Project Discovery:** List all accessible Jira projects with beautifully aligned terminal tables.
* **Issue Tracking:** Search and list Jira issues associated with specific projects using optimized JQL queries.
* **Deep Dives:** Describe specific Jira tickets to view summaries, statuses, assignees, and priorities at a glance.
* **Secure Authentication:** Safely stores and uses Atlassian API tokens to interact with Jira REST API v3.
* **Built for Speed:** Powered by Go and Cobra, offering fast execution and intuitive command aliases.

---

## 🚀 Installation

### Option 1: Build from Source
Ensure you have [Go](https://go.dev/doc/install) installed (1.18 or later recommended).

```bash
git clone [https://github.com/aniruddha-sinha/jiraffe.git](https://github.com/aniruddha-sinha/jiraffe.git)
cd jiraffe
go build -o jiraffe main.go

# Optional: Move to your local bin for global access
mv jiraffe /usr/local/bin/