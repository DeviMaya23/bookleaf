## 1. Backend — Repository Interface and Params

- [x] 1.1 Add `SearchLabels bool` field to `ListImagesParams` in `backend/internal/usecase/image_pagination.go`
- [x] 1.2 Add `searchLabels bool` parameter after `name *string` in the `ImageRepository.List` signature in `backend/internal/usecase/image_repository.go`

## 2. Backend — Repository Implementation

- [x] 2.1 Update `imageRepository.List` in `backend/internal/repository/image_repository.go`: when `name != nil && *name != ""` and `searchLabels == true`, replace the plain `ILIKE` clause with `(images.title ILIKE ? OR EXISTS (SELECT 1 FROM image_labels WHERE image_id = images.id AND label ILIKE ?))`, passing the same `%<term>%` value for both placeholders; otherwise keep the existing plain `ILIKE` clause
- [x] 2.2 Add `AND score >= 0.75` to the EXISTS subquery added in 2.1; define a local unexported constant `labelSearchScoreThreshold float32 = 0.75` in `image_repository.go` and use it as the query argument rather than a bare literal

## 3. Backend — Usecase

- [x] 3.1 Update `imageUsecase.ListImages` in `backend/internal/usecase/image_usecase.go` to pass `params.SearchLabels` through to the repository `List` call

## 4. Backend — Handler

- [x] 4.1 Update the `GET /images` handler in `backend/internal/handler/image.go` to parse `search_labels` from query params (`c.QueryParam("search_labels") == "true"`) and set it on `ListImagesParams.SearchLabels`

## 5. Backend — Unit Tests

- [x] 5.1 Add unit test `TestImageUsecase_ListImages_PassesSearchLabelsToRepository` in `backend/internal/usecase/image_usecase_test.go`: call `ListImages` with `SearchLabels: true`, assert the mock repo receives `searchLabels = true`

## 6. Backend — Bruno

- [x] 6.1 Add `~search_labels: true` as a disabled query param to `bruno/images/list-images.bru`

## 7. Backend — Lint

- [x] 7.1 Run `golangci-lint run ./...` from `backend/` and fix any issues

## 8. Frontend — Controls and Filter Sections

- [x] 8.1 Add `'labelSearch'` to the `FilterSection` union type in `frontend/src/features/gallery/hooks/useGalleryControls.ts`
- [x] 8.2 Update `filterSectionsForViewType` to accept a second `visionEnabled: boolean` parameter and include `'labelSearch'` in the result for `'all'` and `'unsorted'` views when `visionEnabled` is true
- [x] 8.3 Add `searchLabels: boolean` state (defaulting to `false`) to `useGalleryControls`; accept `visionEnabled: boolean` as a new parameter; reset `searchLabels` to `false` in the view-change `useEffect`; expose `searchLabels` and `setSearchLabels` in the return value

## 9. Frontend — AppLayout

- [x] 9.1 Add a `getMe` query to `frontend/src/app-shell/AppLayout.tsx` (same pattern as `useVisionSuggestion`: `useQuery({ queryKey: ['me'], queryFn: () => getMe(getToken) })`)
- [x] 9.2 Pass `me?.vision_enabled ?? false` as the `visionEnabled` argument to `useGalleryControls`
- [x] 9.3 Forward `searchLabels` and `setSearchLabels` from gallery controls down to `ImageGrid` (add to the props passed alongside `debouncedSearchTerm`, `filterTagIds`, etc.)

## 10. Frontend — GalleryToolbar

- [x] 10.1 Add `searchLabels: boolean` and `setSearchLabels: (v: boolean) => void` to the `GalleryToolbarProps` destructure from `controls`
- [x] 10.2 Render a `DropdownMenuCheckboxItem` labelled "Search in AI labels" inside the filter dropdown when `filterSections.includes('labelSearch')`, bound to `searchLabels`/`setSearchLabels`

## 11. Frontend — API Layer

- [x] 11.1 Add optional `searchLabels?: boolean` parameter to `getImages` and `getAllImages` in `frontend/src/lib/images.ts`; when truthy, append `search_labels=true` to the `URLSearchParams`

## 12. Frontend — useGalleryImages

- [x] 12.1 Add `searchLabels: boolean` to `GalleryQueryParams` and `UseGalleryImagesParams` in `frontend/src/features/gallery/hooks/useGalleryImages.ts`
- [x] 12.2 Include `searchLabels` in `queryKeyFor` for `'all'` and `'unsorted'` cases
- [x] 12.3 Pass `searchLabels` to `getAllImages` and `getImages` in `fetcherFor`

## 13. Frontend — ImageGrid

- [x] 13.1 Add `searchLabels?: boolean` prop to `ImageGrid` and forward it to `useGalleryImages`

## 14. Frontend — Build and Lint

- [x] 14.1 Run `npm run build` from `frontend/` and fix any type errors
- [x] 14.2 Run `npm run lint` from `frontend/` and fix any lint issues
