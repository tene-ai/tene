// Design Ref: docs/02-design/features/blog-tts.design.md §4.1
// Pure utilities for extracting speakable text chunks from the rendered
// article DOM and splitting long text into ~160 char sentence-aligned
// chunks (Chrome 15-second utterance cutoff workaround).

export interface TTSChunk {
  text: string;
  blockIndex: number;
  blockEl: HTMLElement;
}

/**
 * Split long text into sentence-aligned chunks of ≤ maxLen characters.
 * Workaround for the Chrome bug that silently stops audio on utterances
 * longer than ~15 seconds (chromium #679437).
 */
export function chunkText(text: string, maxLen = 160): string[] {
  const re = new RegExp(
    `[\\s\\S]{1,${maxLen}}[.!?,](?=\\s|$)|[\\s\\S]{1,${maxLen}}`,
    "g",
  );
  const matches = text.match(re);
  return matches && matches.length > 0 ? matches : [text];
}

const BLOCK_SELECTOR = "p, h2, h3, h4, h5, li, blockquote";

/**
 * Extract TTS-readable chunks from the rendered article DOM.
 * Each chunk carries a reference to its containing block element so the
 * scroll-sync hook can call scrollIntoView() on paragraph transitions.
 */
export function extractChunks(articleEl: HTMLElement): TTSChunk[] {
  const blocks = Array.from(
    articleEl.querySelectorAll<HTMLElement>(BLOCK_SELECTOR),
  )
    // Skip block-level code (<pre> wraps the shiki output). Inline <code>
    // inside a <p> stays — textContent will flatten it.
    .filter((el) => !el.closest("pre"))
    .filter((el) => !isEmpty(el));

  const chunks: TTSChunk[] = [];
  blocks.forEach((el, blockIndex) => {
    const text = readableText(el);
    if (!text) return;
    for (const t of chunkText(text, 160)) {
      chunks.push({ text: t, blockIndex, blockEl: el });
    }
  });
  return chunks;
}

/**
 * textContent minus the rehypeAutolinkHeadings anchor and any aria-hidden
 * decorative children, with whitespace normalized.
 */
function readableText(el: HTMLElement): string {
  const clone = el.cloneNode(true) as HTMLElement;
  clone
    .querySelectorAll('.heading-anchor, [aria-hidden="true"]')
    .forEach((n) => n.remove());
  return (clone.textContent ?? "").replace(/\s+/g, " ").trim();
}

function isEmpty(el: HTMLElement): boolean {
  return !(el.textContent ?? "").trim();
}
