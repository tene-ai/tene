// Design Ref: §2.1 T1-3 — MDX component map used by next-mdx-remote compileMDX.
// Shared styling for all blog articles so layout consistency is a compile-time
// property, not a hand-waved guideline.
import type { MDXComponents } from "mdx/types";
import Link from "next/link";
import { CopyCommand } from "@/components/copy-command";
import { Callout } from "@/components/mdx/callout";
import { GifEmbed } from "@/components/mdx/gif-embed";
import { VideoEmbed } from "@/components/mdx/video-embed";
import { CodeBlockWrapper } from "@/components/blog/code-block-wrapper";
import { TrackedLink } from "@/components/blog/tracked-link";

export function useMDXComponents(
  components: MDXComponents,
  context?: { slug?: string },
): MDXComponents {
  const slug = context?.slug;

  return {
    h1: ({ children, ...props }) => (
      <h1
        className="mt-10 mb-6 text-3xl font-bold tracking-tight sm:text-4xl"
        {...props}
      >
        {children}
      </h1>
    ),
    h2: ({ children, ...props }) => (
      <h2
        className="group mt-12 mb-4 text-2xl font-bold tracking-tight sm:text-3xl scroll-mt-24"
        {...props}
      >
        {children}
      </h2>
    ),
    h3: ({ children, ...props }) => (
      <h3
        className="mt-8 mb-3 text-xl font-semibold scroll-mt-24"
        {...props}
      >
        {children}
      </h3>
    ),
    h4: ({ children, ...props }) => (
      <h4 className="mt-6 mb-2 text-lg font-semibold scroll-mt-24" {...props}>
        {children}
      </h4>
    ),
    p: ({ children, ...props }) => (
      <p className="my-4 leading-relaxed text-foreground/90" {...props}>
        {children}
      </p>
    ),
    a: ({ href, children, ...props }) => {
      if (!href) return <span>{children}</span>;
      const isExternal = /^https?:\/\//.test(href);
      if (isExternal) {
        // Tracked external link — fires blog_external_link analytics event.
        return (
          <TrackedLink href={href} slug={slug} {...props}>
            {children}
          </TrackedLink>
        );
      }
      return (
        <Link
          href={href}
          className="text-accent underline underline-offset-4 hover:text-accent-dim"
          {...props}
        >
          {children}
        </Link>
      );
    },
    ul: ({ children, ...props }) => (
      <ul className="my-4 ml-5 list-disc space-y-2" {...props}>
        {children}
      </ul>
    ),
    ol: ({ children, ...props }) => (
      <ol className="my-4 ml-5 list-decimal space-y-2" {...props}>
        {children}
      </ol>
    ),
    li: ({ children, ...props }) => (
      <li className="leading-relaxed" {...props}>
        {children}
      </li>
    ),
    blockquote: ({ children, ...props }) => (
      <blockquote
        className="my-6 border-l-4 border-accent/50 bg-surface/60 pl-4 py-2 italic text-muted"
        {...props}
      >
        {children}
      </blockquote>
    ),
    // Inline code vs code block — rehype-shiki renders blocks with className.
    code: ({ children, className, ...props }) => {
      if (!className) {
        // Inline code
        return (
          <code
            className="rounded bg-surface px-1.5 py-0.5 font-mono text-sm text-accent"
            {...props}
          >
            {children}
          </code>
        );
      }
      // Block code — shiki adds class like "language-bash shiki ..."
      return (
        <code className={className} {...props}>
          {children}
        </code>
      );
    },
    // pre is wrapped with a client-side CodeBlockWrapper that adds copy button
    // + fires blog_copy_code analytics event on click.
    pre: (props) => <CodeBlockWrapper slug={slug} {...props} />,
    img: ({ src, alt }) => {
      if (!src || typeof src !== "string") return null;
      // Use plain <img> for GIFs (Next/Image chokes on animated GIFs) and for
      // remote URLs. Image component is preferred for static PNG/JPEG but blog
      // author mixes both freely — keep it simple.
      return (
        <img
          src={src}
          alt={alt ?? ""}
          loading="lazy"
          className="my-6 h-auto w-full rounded-lg border border-border"
        />
      );
    },
    table: ({ children, ...props }) => (
      <div className="my-6 -mx-4 overflow-x-auto px-4 sm:mx-0 sm:px-0">
        <table className="w-full text-sm" {...props}>
          {children}
        </table>
      </div>
    ),
    thead: ({ children, ...props }) => <thead {...props}>{children}</thead>,
    th: ({ children, ...props }) => (
      <th
        className="border-b border-border py-2 px-3 text-left font-medium"
        {...props}
      >
        {children}
      </th>
    ),
    td: ({ children, ...props }) => (
      <td className="border-b border-border/50 py-2 px-3" {...props}>
        {children}
      </td>
    ),
    hr: () => <hr className="my-10 border-border/50" />,
    // Custom reusable MDX components
    CopyCommand,
    Callout,
    GifEmbed,
    VideoEmbed,
    ...components,
  };
}
