# Database migrations

SQL migrations live in this directory. Run them with:

```bash
gofreight migrate
gofreight migrate:status
```

Create a new migration:

```bash
gofreight make:migration create_posts_table
```

Files ending in _down.sql are used for rollbacks.
