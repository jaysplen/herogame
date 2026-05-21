import { useEffect, useState } from "react";
import { Image as KonvaImage } from "react-konva";

/**
 * Watercolor map background — loads `/map-watercolor.svg` (served from
 * frontend/public) and renders it as a Konva.Image at z=0.
 *
 * Falls back to nothing (transparent) while loading or if the asset is
 * missing, so the stage still renders.
 */
export function MapBackground({
  width,
  height,
  src = "/map-watercolor.svg",
}: {
  width: number;
  height: number;
  src?: string;
}) {
  const [img, setImg] = useState<HTMLImageElement | null>(null);

  useEffect(() => {
    const el = new window.Image();
    el.crossOrigin = "anonymous";
    el.src = src;
    let cancelled = false;
    el.onload = () => {
      if (!cancelled) setImg(el);
    };
    el.onerror = () => {
      if (!cancelled) setImg(null);
    };
    return () => {
      cancelled = true;
    };
  }, [src]);

  if (!img) return null;
  return (
    <KonvaImage
      image={img}
      x={0}
      y={0}
      width={width}
      height={height}
      listening={false}
    />
  );
}
