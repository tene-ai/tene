"use client";

import { useEffect, useRef, useCallback } from "react";

const DOT_SPACING = 24;
const DOT_RADIUS = 1;
const MOUSE_RADIUS = 120;
const PUSH_STRENGTH = 8;
const DOT_COLOR = "#2a2a2a";

export function InteractiveGrid() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const mouseRef = useRef({ x: -1000, y: -1000 });
  const frameRef = useRef<number>(0);
  const dotsRef = useRef<{ baseX: number; baseY: number }[]>([]);
  const sizeRef = useRef({ w: 0, h: 0 });

  const initDots = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const dpr = window.devicePixelRatio || 1;
    const w = window.innerWidth;
    const h = window.innerHeight;

    canvas.width = w * dpr;
    canvas.height = h * dpr;
    canvas.style.width = `${w}px`;
    canvas.style.height = `${h}px`;

    sizeRef.current = { w, h };

    const dots: { baseX: number; baseY: number }[] = [];
    for (let x = DOT_SPACING / 2; x < w; x += DOT_SPACING) {
      for (let y = DOT_SPACING / 2; y < h; y += DOT_SPACING) {
        dots.push({ baseX: x, baseY: y });
      }
    }
    dotsRef.current = dots;
  }, []);

  const draw = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const dpr = window.devicePixelRatio || 1;
    const { w, h } = sizeRef.current;
    const mx = mouseRef.current.x;
    const my = mouseRef.current.y;

    ctx.clearRect(0, 0, w * dpr, h * dpr);
    ctx.save();
    ctx.scale(dpr, dpr);

    for (const dot of dotsRef.current) {
      const dx = dot.baseX - mx;
      const dy = dot.baseY - my;
      const dist = Math.sqrt(dx * dx + dy * dy);

      let drawX = dot.baseX;
      let drawY = dot.baseY;
      let radius = DOT_RADIUS;
      let color = DOT_COLOR;

      if (dist < MOUSE_RADIUS) {
        const factor = 1 - dist / MOUSE_RADIUS;
        const push = factor * factor * PUSH_STRENGTH;

        const angle = Math.atan2(dy, dx);
        drawX = dot.baseX + Math.cos(angle) * push;
        drawY = dot.baseY + Math.sin(angle) * push;

        radius = DOT_RADIUS + factor * 1.5;

        const g = Math.round(42 + factor * (255 - 42));
        const r = Math.round(42 + factor * (0 - 42));
        const b = Math.round(42 + factor * (136 - 42));
        const a = 0.4 + factor * 0.6;
        color = `rgba(${Math.max(0, r)}, ${g}, ${b}, ${a})`;
      }

      ctx.beginPath();
      ctx.arc(drawX, drawY, radius, 0, Math.PI * 2);
      ctx.fillStyle = color;
      ctx.fill();
    }

    ctx.restore();
    frameRef.current = requestAnimationFrame(draw);
  }, []);

  useEffect(() => {
    if (window.matchMedia("(max-width: 640px)").matches) return;
    if ("ontouchstart" in window) return;

    initDots();
    frameRef.current = requestAnimationFrame(draw);

    const handleMouseMove = (e: MouseEvent) => {
      mouseRef.current = { x: e.clientX, y: e.clientY };
    };

    const handleResize = () => {
      initDots();
    };

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("resize", handleResize);

    return () => {
      cancelAnimationFrame(frameRef.current);
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("resize", handleResize);
    };
  }, [initDots, draw]);

  return (
    <>
      <canvas
        ref={canvasRef}
        className="pointer-events-none fixed inset-0 z-0 hidden sm:block"
      />
      {/* Mobile: static CSS dot grid fallback */}
      <div className="dot-grid sm:hidden" />
    </>
  );
}
