interface WordPopoverProps {
  word: string | null
  onClose: () => void
}

export default function WordPopover({ word }: WordPopoverProps) {
  if (word === null) return null

  return <div>{word}</div>
}
