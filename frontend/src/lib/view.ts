export type AppView =
  | { type: 'all' }
  | { type: 'unsorted' }
  | { type: 'trash' }
  | { type: 'folder'; id: string }
