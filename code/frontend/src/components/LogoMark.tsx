/**
 * Đèo's mark: a saddle between two peaks — the shape of a mountain pass,
 * which is what "đèo" means and what these bikes are bought for. Drawn in
 * currentColor; public/favicon.svg is the same mark.
 */
export default function LogoMark({ className = "" }: { className?: string }) {
  return (
    <svg viewBox="0 0 32 32" className={className} aria-hidden="true" fill="currentColor">
      <path d="M2 26 L11 8 L16.5 18 L20 12 L30 26 Z" />
    </svg>
  );
}
