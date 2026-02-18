Pomodoro + Notes API — Go Project
This is a dual-purpose Go project combining a Pomodoro productivity timer and a RESTful Notes/Todo API, both built from scratch using only Go's standard library (no heavy frameworks).

->Pomodoro Timer:
A command-line productivity timer based on the Pomodoro Technique — a time management method that breaks work into focused intervals separated by short breaks. The timer runs fully in the terminal, displays a live countdown, sends native desktop notifications on Windows, macOS, and Linux, and automatically cycles through work sessions, short breaks, and long breaks. Everything is configurable via flags, so you can adjust session length, break durations, and the number of rounds to fit your workflow.

->Notes & Todo REST API:
A lightweight HTTP API server for managing personal notes and todos. It supports full CRUD operations (create, read, update, delete) and includes a complete authentication system — users can register, log in, and receive a secure token that protects their data. Each user only sees their own notes. Notes support titles, body text, priority levels (low / medium / high), tags, and a done/undone status.
