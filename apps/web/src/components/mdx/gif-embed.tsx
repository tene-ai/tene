// Design Ref: §2.5 — GIF/image with caption. Plain <img> since Next/Image
// has issues with animated GIFs in Turbopack.
type Props = {
  src: string;
  alt: string;
  caption?: string;
  width?: number;
  height?: number;
};

export function GifEmbed({ src, alt, caption, width = 800, height = 450 }: Props) {
  return (
    <figure className="my-8">
      <img
        src={src}
        alt={alt}
        width={width}
        height={height}
        loading="lazy"
        className="w-full rounded-lg border border-border"
      />
      {caption && (
        <figcaption className="mt-2 text-center text-sm text-muted">
          {caption}
        </figcaption>
      )}
    </figure>
  );
}
