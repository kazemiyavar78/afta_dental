import { useEffect, useRef } from 'react';
import { TOOTH_CHART_SVG } from './toothChartSvg';
import {
  quadrantToUniversal,
  universalDisplayNumber,
  universalToQuadrant,
  type ToothQuadrantSelection,
} from './toothQuadrant';

const SELECTED_FILL = '#3D8BFD';
const DEFAULT_FILL = '#FFFFFF';
const SELECTED_LABEL = '#0B5ED7';
const DEFAULT_LABEL = '#334155';

/**
 * موقعیت افقی نهایی برچسب‌ها (داخل قوس، نزدیک لبه دندان).
 * مقادیر مطلق‌اند تا با remount / Strict Mode تکراری جابه‌جا نشوند.
 */
const LABEL_X: Record<number, number> = {
  1: 105,
  2: 108,
  3: 110,
  4: 120,
  5: 140,
  6: 160,
  7: 175,
  8: 210,
  9: 250,
  10: 280,
  11: 292,
  12: 309,
  13: 322,
  14: 333,
  15: 337,
  16: 335,
  17: 346,
  18: 347,
  19: 345,
  20: 326,
  21: 309,
  22: 298,
  23: 278,
  24: 253,
  25: 211,
  26: 183,
  27: 159,
  28: 145,
  29: 130,
  30: 110,
  31: 109,
  32: 110,
};

export type ToothSelectionProps = {
  /** شماره دندان انتخاب‌شده داخل ربع (۱–۸) */
  selectedNumber?: number | null;
  /** جهت ربع انتخاب‌شده (۱–۴) */
  selectedDirection?: number | null;
  /** پس از کلیک روی دندان فراخوانی می‌شود */
  onSelect?: (selection: ToothQuadrantSelection) => void;
};

/**
 * نمودار تعاملی انتخاب دندان با شماره‌گذاری چهارربعی (۸×۴).
 * لایه خطوط روی دندان کلیک را مسدود نمی‌کند؛ برچسب‌ها با مختصات مطلق تنظیم می‌شوند.
 */
export function ToothSelection({
  selectedNumber = null,
  selectedDirection = null,
  onSelect,
}: ToothSelectionProps) {
  const rootRef = useRef<HTMLDivElement>(null);
  const onSelectRef = useRef(onSelect);
  onSelectRef.current = onSelect;

  const selectedUniversal =
    selectedNumber != null && selectedDirection != null
      ? quadrantToUniversal(selectedDirection, selectedNumber)
      : null;

  /** آماده‌سازی SVG بعد از mount: موقعیت و متن برچسب‌های ۱–۸ */
  useEffect(() => {
    const root = rootRef.current;
    if (!root) return;

    const outlines = root.querySelector('#adult-outlines');
    if (outlines instanceof SVGElement) {
      outlines.style.pointerEvents = 'none';
    }
    const labelsGroup = root.querySelector('#toothLabels');
    if (labelsGroup instanceof SVGElement) {
      labelsGroup.style.pointerEvents = 'none';
    }

    for (let universal = 1; universal <= 32; universal += 1) {
      const el = root.querySelector(`#lbl${universal}`);
      if (!(el instanceof SVGElement)) continue;

      const display = universalDisplayNumber(universal);
      if (display != null) {
        el.textContent = String(display);
      }

      const x = LABEL_X[universal];
      if (x == null) continue;
      const current = el.getAttribute('transform') ?? '';
      const matrixMatch = current.match(/matrix\(([^)]+)\)/);
      if (!matrixMatch) continue;
      const parts = matrixMatch[1].trim().split(/\s+/);
      if (parts.length < 6) continue;
      parts[4] = String(x);
      el.setAttribute('transform', `matrix(${parts.join(' ')})`);
    }

    const svg = root.querySelector('svg');
    if (svg) {
      svg.style.overflow = 'visible';
    }
  }, []);

  /** کلیک روی دندان — برگرداندن جهت و شماره ۱–۸ */
  useEffect(() => {
    const root = rootRef.current;
    if (!root) return;

    const handleClick = (event: MouseEvent) => {
      const hit = document
        .elementsFromPoint(event.clientX, event.clientY)
        .find((el) => el.hasAttribute('data-key'));
      if (!hit) return;

      const universal = Number(hit.getAttribute('data-key'));
      const selection = universalToQuadrant(universal);
      if (!selection) return;

      hit.classList.remove('is-pulse');
      void (hit as HTMLElement).getBoundingClientRect?.();
      hit.classList.add('is-pulse');
      onSelectRef.current?.(selection);
    };

    root.addEventListener('click', handleClick);
    return () => root.removeEventListener('click', handleClick);
  }, []);

  /** رنگ دندان و برچسب انتخاب‌شده */
  useEffect(() => {
    const root = rootRef.current;
    if (!root) return;

    root.querySelectorAll('#Spots [data-key]').forEach((el) => {
      const num = Number(el.getAttribute('data-key'));
      const isSelected = selectedUniversal != null && num === selectedUniversal;
      el.setAttribute('fill', isSelected ? SELECTED_FILL : DEFAULT_FILL);
      el.classList.toggle('is-selected', isSelected);
    });

    root.querySelectorAll('#toothLabels text').forEach((el) => {
      const id = el.getAttribute('id') ?? '';
      const num = Number(id.replace(/^lbl/i, ''));
      const isSelected = selectedUniversal != null && num === selectedUniversal;
      el.setAttribute('fill', isSelected ? SELECTED_LABEL : DEFAULT_LABEL);
      el.classList.toggle('is-selected-label', isSelected);
    });
  }, [selectedUniversal]);

  return (
    <div className="tooth-chart-wrap">
      <div className="tooth-chart-stage">
        <div
          ref={rootRef}
          className="tooth-chart"
          dangerouslySetInnerHTML={{ __html: TOOTH_CHART_SVG }}
        />
      </div>
      <style>{`
        .tooth-chart-wrap {
          display: flex;
          justify-content: center;
          padding: 8px 4px 4px;
        }

        .tooth-chart-stage {
          position: relative;
          width: 480px;
          max-width: 100%;
          border-radius: 20px;
          padding: 16px 24px;
          background:
            radial-gradient(ellipse 70% 45% at 50% 18%, rgba(61, 139, 253, 0.12), transparent 60%),
            radial-gradient(ellipse 70% 45% at 50% 88%, rgba(61, 139, 253, 0.10), transparent 60%),
            linear-gradient(180deg, #f8fafc 0%, #eef4ff 100%);
          box-shadow:
            inset 0 1px 0 rgba(255, 255, 255, 0.8),
            0 10px 28px rgba(15, 23, 42, 0.08);
          animation: tooth-stage-in 420ms ease-out;
        }

        .tooth-chart {
          width: 100%;
        }

        .tooth-chart svg {
          width: 100%;
          height: auto;
          display: block;
          overflow: visible;
        }

        #adult-outlines,
        #toothLabels {
          pointer-events: none !important;
        }

        #toothLabels text {
          transition: fill 0.25s ease, opacity 0.25s ease;
          opacity: 0.9;
          user-select: none;
        }

        #toothLabels text.is-selected-label {
          opacity: 1;
          font-weight: 700;
          filter: drop-shadow(0 0 4px rgba(11, 94, 215, 0.35));
        }

        #Spots polygon,
        #Spots path {
          cursor: pointer;
          pointer-events: auto !important;
          transform-box: fill-box;
          transform-origin: center;
          transition:
            fill 0.28s ease,
            stroke 0.28s ease,
            filter 0.28s ease,
            transform 0.28s cubic-bezier(0.22, 1, 0.36, 1);
        }

        #Spots polygon:hover,
        #Spots path:hover {
          fill: #dbe8ff !important;
          transform: scale(1.045);
          filter: drop-shadow(0 4px 8px rgba(61, 139, 253, 0.22));
        }

        #Spots polygon.is-selected,
        #Spots path.is-selected {
          fill: ${SELECTED_FILL} !important;
          stroke: #1d4ed8;
          stroke-width: 1.5;
          filter: drop-shadow(0 0 10px rgba(61, 139, 253, 0.55));
          transform: scale(1.06);
          animation: tooth-selected-glow 1.6s ease-in-out infinite;
        }

        #Spots polygon.is-selected:hover,
        #Spots path.is-selected:hover {
          fill: ${SELECTED_FILL} !important;
        }

        #Spots polygon.is-pulse,
        #Spots path.is-pulse {
          animation: tooth-click-pulse 420ms ease-out;
        }

        @keyframes tooth-stage-in {
          from { opacity: 0; transform: translateY(8px) scale(0.98); }
          to { opacity: 1; transform: translateY(0) scale(1); }
        }

        @keyframes tooth-selected-glow {
          0%, 100% { filter: drop-shadow(0 0 8px rgba(61, 139, 253, 0.45)); }
          50% { filter: drop-shadow(0 0 14px rgba(61, 139, 253, 0.75)); }
        }

        @keyframes tooth-click-pulse {
          0% { transform: scale(1); }
          40% { transform: scale(1.12); }
          100% { transform: scale(1.06); }
        }
      `}</style>
    </div>
  );
}
