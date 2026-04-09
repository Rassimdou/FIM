# FIM
FIM is a distributed File Integrity Monitoring project. The goal is to run lightweight agents on monitored machines, detect filesystem activity, and send structured events to a central service for analysis, observability, and alerting.

This repository is still in progress. 
## Why This Project Exists

File Integrity Monitoring helps answer simple but important questions:

- What changed on a system?
- When did it change?
- Which file was affected?
- Was it a create, modify, delete, permission, or ownership change?
- Does the change look expected or suspicious?

The long-term idea for this project is to detect those events on endpoints, centralize them over gRPC, enrich them with hashes and metadata, and make them easy to inspect through logs, dashboards, and alerts.

## Architecture

At a high level, the system is designed around three parts.

### 1. Endpoint Agents

Agents run on monitored machines and watch the local filesystem.

- Linux agent: `fsnotify` / inotify-based watcher
- Windows agent: planned around `ReadDirectoryChangesW`
- macOS agent: planned around `FSEvents` or `kqueue`

Each agent is responsible for:

- loading local watch configuration
- subscribing to filesystem changes on selected paths
- converting native watcher events into a shared protobuf event format
- sending events to the central server over a gRPC stream

### 2. Central Server

The server is intended to receive event streams from agents and process them centrally.

Planned responsibilities include:

- accepting agent event streams over gRPC
- comparing events against a baseline or expected state
- enriching events with additional information such as hashes or permission diffs
- forwarding structured records into the observability pipeline

### 3. Observability Stack

The observability side is planned to make the event stream usable in practice.

- Loki for structured log ingestion
- Grafana for dashboards and investigation
- Alertmanager for notifications such as email or webhook alerts

The intended event flow is:

`agent -> gRPC stream (mTLS encrypted) -> server/event engine -> structured logs -> Loki -> Grafana/Alertmanager`

## Agent Stability
The endpoint agent is built with robust stability features to prevent crashing endpoints:
- **Baseline Scanning:** Establishes initial hashes on startup and reconnects.
- **Max File Size:** Avoids memory and CPU starvation by skipping hashes for extremely large files (e.g. databases, ISOs).
- **Race Condition Retries:** Automatically handles temporary file locks (especially on Windows) allowing applications to finish writing data before attempting to compute a file's hash.

## Security
This project uses **Mutual TLS (mTLS)** for the gRPC stream. Not only is the data encrypted, but the central server will strictly reject any agent that does not present a cryptographic certificate signed by our own local Certificate Authority (CA).

See the [Security Documentation](docs/security.md) for more details and instructions on generating your local development certificates.

