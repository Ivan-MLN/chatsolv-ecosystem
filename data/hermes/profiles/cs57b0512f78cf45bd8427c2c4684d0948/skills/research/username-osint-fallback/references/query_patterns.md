# Username OSINT Search Queries & Fallback Strategy

When CLI username lookup tools like `sherlock` fail due to environment or dependency issues, use the following structured queries with Serper API or web search:

## Query Patterns

```python
queries = [
    '"{username}"',
    '"{alt_username}"',
    'site:github.com "{username}"',
    'site:linkedin.com/in "{username}"',
    'site:instagram.com "{username}"',
    'site:tiktok.com "{username}"',
    'site:x.com "{username}"'
]
```

## Verification Checklist

1. Check if the handle matches exact username syntax (with or without underscores).
2. Check for public comments or issue contributions (e.g. `whatsmeow`, GitHub repos).
3. Validate location or bio details across returned matches before reporting.
