import { Plus } from 'lucide-react'

interface FloatingUploadButtonProps {
  onClick: () => void
}

export default function FloatingUploadButton({ onClick }: FloatingUploadButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label="Upload image"
      className="sm:hidden fixed bottom-5 right-5 z-20 w-12 h-12 rounded-full bg-primary text-primary-foreground shadow-lg flex items-center justify-center hover:bg-primary/90"
    >
      <Plus className="w-5 h-5" />
    </button>
  )
}
