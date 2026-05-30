## 1. Backend — Return Error on Thumbnail Failure

- [x] 1.1 In `CompleteUpload`, replace `return result, nil` on `prepareThumbnail` failure with `return nil, err`

## 2. Tests — Update Unit Tests

- [x] 2.1 Update the `GetObject failure` scenario test to assert `CompleteUpload` returns an error (not a warning)
- [x] 2.2 Add or update `Generate failure` scenario test to assert `CompleteUpload` returns an error
