import { useState } from "react"
import { Film, Loader2, Play, Trash2 } from "lucide-react"
import { toast } from "sonner"
import type { Item } from "@/lib/api"
import { api } from "@/lib/api"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"

function confidenceVariant(conf: number): "default" | "secondary" | "destructive" {
  if (conf < 0.5) return "destructive"
  if (conf < 0.7) return "secondary"
  return "default"
}

export function ItemCard({
  item,
  onAccepted,
  onRejected,
}: {
  item: Item
  onAccepted: (id: string) => void
  onRejected: (id: string) => void
}) {
  const [artist, setArtist] = useState(item.inferredArtist || item.artistGuess)
  const [title, setTitle] = useState(item.inferredTitle || item.rawTitle)
  const [showVideo, setShowVideo] = useState(false)
  const [busy, setBusy] = useState(false)

  async function handleAccept() {
    if (!artist.trim() || !title.trim()) return
    setBusy(true)
    try {
      await api.accept(item.id, artist.trim(), title.trim())
      toast.success(`Accepted "${title.trim()}"`)
      onAccepted(item.id)
    } catch (e) {
      toast.error(`Failed to accept: ${(e as Error).message}`)
      setBusy(false)
    }
  }

  async function handleReject() {
    setBusy(true)
    try {
      await api.reject(item.id)
      toast.success(`Deleted "${item.filename}"`)
      onRejected(item.id)
    } catch (e) {
      toast.error(`Failed to delete: ${(e as Error).message}`)
      setBusy(false)
    }
  }

  return (
    <Card className="overflow-hidden">
      <CardHeader className="gap-2">
        <div className="relative aspect-video overflow-hidden rounded-md bg-muted">
          {showVideo ? (
            // eslint-disable-next-line jsx-a11y/media-has-caption
            <video
              className="h-full w-full"
              src={api.streamUrl(item.id)}
              controls
              autoPlay
            />
          ) : item.hasThumbnail ? (
            <button
              type="button"
              className="group relative h-full w-full cursor-pointer"
              onClick={() => setShowVideo(true)}
            >
              <img
                src={api.thumbnailUrl(item.id)}
                alt=""
                className="h-full w-full object-cover"
              />
              <span className="absolute inset-0 flex items-center justify-center bg-black/0 transition-colors group-hover:bg-black/30">
                <Play className="h-10 w-10 text-white opacity-0 transition-opacity group-hover:opacity-100" />
              </span>
            </button>
          ) : (
            <button
              type="button"
              className="flex h-full w-full items-center justify-center gap-2 text-muted-foreground"
              onClick={() => setShowVideo(true)}
            >
              <Film className="h-8 w-8" />
              <span className="text-sm">Preview</span>
            </button>
          )}
        </div>

        <div className="flex flex-wrap gap-1.5">
          <Badge variant="outline">{item.category}</Badge>
          {item.inferredIsFullSet && <Badge variant="outline">Full set</Badge>}
          <Badge variant={confidenceVariant(item.inferredArtistConfidence)}>
            artist {item.inferredArtistConfidence.toFixed(2)}
          </Badge>
          <Badge variant={confidenceVariant(item.inferredTitleConfidence)}>
            title {item.inferredTitleConfidence.toFixed(2)}
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="space-y-3">
        <div className="space-y-1">
          <Label htmlFor={`artist-${item.id}`}>Artist</Label>
          <Input
            id={`artist-${item.id}`}
            value={artist}
            onChange={(e) => setArtist(e.target.value)}
            disabled={busy}
          />
        </div>
        <div className="space-y-1">
          <Label htmlFor={`title-${item.id}`}>Title</Label>
          <Input
            id={`title-${item.id}`}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            disabled={busy}
          />
        </div>

        <details className="text-sm text-muted-foreground">
          <summary className="cursor-pointer select-none">
            Original metadata
          </summary>
          <div className="mt-2 space-y-1">
            <p className="break-words">
              <span className="font-medium">Raw title:</span> {item.rawTitle}
            </p>
            {item.uploader && (
              <p>
                <span className="font-medium">Uploader:</span> {item.uploader}
              </p>
            )}
            {item.description && (
              <Textarea
                readOnly
                value={item.description}
                className="h-24 resize-none text-xs"
              />
            )}
          </div>
        </details>
      </CardContent>

      <CardFooter className="flex gap-2">
        <Button
          className="flex-1"
          onClick={handleAccept}
          disabled={busy || !artist.trim() || !title.trim()}
        >
          {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : "Accept"}
        </Button>

        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button variant="destructive" size="icon" disabled={busy}>
              <Trash2 className="h-4 w-4" />
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Delete this download?</AlertDialogTitle>
              <AlertDialogDescription>
                "{item.filename}" and its thumbnail/subtitle files will be
                permanently deleted. This can't be undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction onClick={handleReject}>
                Delete
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </CardFooter>
    </Card>
  )
}
