// Design Ref: §2.5 — Self-hosted mp4 or YouTube iframe.
// Note: YouTube iframe requires CSP frame-src https://www.youtube.com. For
// initial blog posts we stick to mp4 self-host to avoid expanding CSP.
type Props = {
  src: string;
  type?: "mp4" | "youtube";
  poster?: string;
  caption?: string;
};

export function VideoEmbed({ src, type = "mp4", poster, caption }: Props) {
  if (type === "youtube") {
    return (
      <figure className="my-8">
        <div className="aspect-video overflow-hidden rounded-lg border border-border">
          <iframe
            src={src}
            title={caption ?? "Video"}
            loading="lazy"
            allow="autoplay; encrypted-media; picture-in-picture"
            allowFullScreen
            className="h-full w-full"
          />
        </div>
        {caption && (
          <figcaption className="mt-2 text-center text-sm text-muted">
            {caption}
          </figcaption>
        )}
      </figure>
    );
  }

  return (
    <figure className="my-8">
      <video
        controls
        poster={poster}
        className="w-full rounded-lg border border-border"
        preload="metadata"
      >
        <source src={src} type="video/mp4" />
        Your browser does not support the video tag.
      </video>
      {caption && (
        <figcaption className="mt-2 text-center text-sm text-muted">
          {caption}
        </figcaption>
      )}
    </figure>
  );
}
