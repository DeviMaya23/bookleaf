import { iconForKey } from '../lib/folderIcons'

interface FolderIconViewProps {
  icon: string | null | undefined
  className?: string
}

/* eslint-disable react-hooks/static-components -- picks among a fixed,
   pre-defined set of icon components by key; not ad hoc creation */
export default function FolderIconView({ icon, className }: FolderIconViewProps) {
  const Icon = iconForKey(icon)
  return <Icon className={className} />
}
/* eslint-enable react-hooks/static-components */
