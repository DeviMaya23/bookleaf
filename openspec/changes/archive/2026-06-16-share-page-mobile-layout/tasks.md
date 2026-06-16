## 1. SharedFolderPanel — mobile layout

- [x] 1.1 Replace outer `flex flex-col h-full w-[280px] border-l flex-shrink-0` with mobile-first classes: `flex flex-col border-b sm:h-full sm:w-[280px] sm:border-b-0 sm:border-l sm:flex-shrink-0`
- [x] 1.2 Strip `flex-1 overflow-y-auto` from the notes section div; replace with `sm:flex-1 sm:overflow-y-auto`

## 2. SharePage — responsive scroll container

- [x] 2.1 Replace `flex-1 flex overflow-hidden` on the middle row with `flex-1 flex flex-col overflow-y-auto sm:flex-row sm:overflow-hidden`
- [x] 2.2 Replace `flex-1 overflow-y-auto p-6` on the gallery container with `p-6 sm:flex-1 sm:overflow-y-auto`
- [x] 2.3 Add `order-first sm:order-last` to `<SharedFolderPanel />` in `SharePage` so it appears above the gallery on mobile while staying last in the DOM

## 3. Verify and build

- [x] 3.1 Manually verify layout at 375px width: panel stacks above gallery, single scroll surface, no horizontal overflow
- [x] 3.2 Manually verify layout at 640px and above: side-by-side layout restored, panel has sticky export footer, gallery scrolls independently
- [x] 3.3 Verify masonry column count at mobile width (should be 1 col at ~375px, images fully visible)
- [x] 3.4 Run `npm run build` and `npm run lint` and fix any issues
