<!-- CONTEXT_START
type: feature-index
project: rancago
CONTEXT_END -->

# Feature Docs Index

Each file is a compact context document for one bounded context.
Read the relevant file before editing feature code.

| Feature | Doc | Description |
|---------|-----|-------------|
| Notification | [notification.md](notification.md) | Real-time notifications + WebSocket + Redis unread counter |
| User | [user.md](user.md) | Auth, OAuth/Socialite, RBAC roles & permissions |
| Document | [document.md](document.md) | Document storage + pgvector similarity search |

## Creating a New Feature

```bash
rancago make:feature MyFeature
```

This scaffolds all hexagonal layers **and** generates `docs/features/my_feature.md` automatically.

## Compact Prompt Technique

Reference the feature doc when working on a feature to provide full context in one line:

```
docs/features/notification.md
Add a `DeleteAll` method that removes all notifications for a given user_id.
```

The context file carries architecture rules, file paths, and output hints — no need to re-explain them every time.
