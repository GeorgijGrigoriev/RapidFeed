# TODO

- [ ] Add DB migration/backfill script to normalize existing feed text fields (decode HTML entities, strip leftover HTML, normalize whitespace).
- [ ] Roll back on-the-fly normalization after migration/backfill is in place.
- [ ] Disable `Force sync` and `Add feed` buttons while waiting for server response, and show a visible loading state/message during the request.
