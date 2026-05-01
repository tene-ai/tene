// Convert frontmatter calendar dates ("YYYY-MM-DD") to the full ISO 8601
// datetime format required by Schema.org datePublished / dateModified and the
// OpenGraph article:published_time / article:modified_time meta tags.
//
// Why this exists: Google Rich Results Test rejects date-only values with
//   - 'datePublished'의 datetime 값이 잘못됨
//   - datetime 속성에 시간대가 누락됨
// (see https://search.google.com/test/rich-results)
//
// Why UTC (Z), not a fixed local offset:
//   tene.sh is a globally-targeted blog. Anchoring at the operator's local
//   timezone (e.g. KST +09:00) bakes a bias into machine-readable surfaces
//   that have no reason to favour any region. UTC is what Google's own
//   documentation samples use, what `sitemap.ts` already emits, and what
//   every search engine normalises against. Pick one neutral anchor and
//   keep it consistent across every emission boundary.
//
// Authors keep writing date-only frontmatter ("2026-05-02") — that stays
// human-readable and locale-agnostic. The conversion to a full datetime
// happens at the emission boundary (this helper) so no future post can
// accidentally ship without a timezone.
export function toIsoDateTime(value: string): string {
  // Already a full datetime (has 'T' followed by digits)? leave alone.
  if (/T\d/.test(value)) return value;
  return `${value}T00:00:00.000Z`;
}
