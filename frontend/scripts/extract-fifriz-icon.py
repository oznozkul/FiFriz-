from __future__ import annotations

import os

import numpy as np
from PIL import Image

SRC = r"C:\Users\alioz\.cursor\projects\c-Users-alioz-Documents-FileAi\assets\c__Users_alioz_AppData_Roaming_Cursor_User_workspaceStorage_7a88f896bce2bdaac75b4fe72014455c_images_WhatsApp_Image_2026-06-27_at_18.04.29-4211d888-1944-41ba-a51f-0b6273e46851.png"
OUT_DIR = os.path.join(os.path.dirname(__file__), "..", "public")


def main() -> None:
    os.makedirs(OUT_DIR, exist_ok=True)

    img = Image.open(SRC).convert("RGBA")
    arr = np.array(img)
    height, width = arr.shape[:2]

    region = arr[:500]
    white = (region[:, :, 0] > 200) & (region[:, :, 1] > 200) & (region[:, :, 2] > 200)
    ys, xs = np.where(white)
    doc_x0, doc_y0, doc_x1, doc_y1 = xs.min(), ys.min(), xs.max(), ys.max()

    pad_left, pad_right, pad_top, pad_bottom = 70, 70, 130, 45
    x0 = max(0, doc_x0 - pad_left)
    x1 = min(width, doc_x1 + pad_right)
    y0 = max(0, doc_y0 - pad_top)
    y1 = min(height, doc_y1 + pad_bottom)

    crop_w, crop_h = x1 - x0, y1 - y0
    side = max(crop_w, crop_h)
    cx, cy = (x0 + x1) // 2, (y0 + y1) // 2
    x0 = max(0, cx - side // 2)
    y0 = max(0, cy - side // 2)
    x1 = min(width, x0 + side)
    y1 = min(height, y0 + side)
    if x1 - x0 < side:
        x0 = max(0, x1 - side)
    if y1 - y0 < side:
        y0 = max(0, y1 - side)

    cropped = arr[y0:y1, x0:x1].copy()

    corners = np.array(
        [
            cropped[0, 0, :3],
            cropped[0, -1, :3],
            cropped[-1, 0, :3],
            cropped[-1, -1, :3],
        ],
        dtype=np.float32,
    )
    bg = np.median(corners, axis=0)

    rgb = cropped[:, :, :3].astype(np.float32)
    alpha = cropped[:, :, 3].astype(np.float32)
    dist = np.sqrt(np.sum((rgb - bg) ** 2, axis=2))

    luminance = rgb.mean(axis=2)
    dark = luminance < 55
    transparent_mask = (dist < 48) | dark
    alpha[transparent_mask] = 0
    fringe = (dist >= 48) & (dist < 70) & ~dark
    alpha[fringe] = np.minimum(alpha[fringe], ((dist[fringe] - 48) / 22) * 255)

    out = cropped.copy()
    out[:, :, 3] = alpha.astype(np.uint8)
    icon_img = Image.fromarray(out)

    icon_path = os.path.join(OUT_DIR, "fifriz-icon.png")
    icon_img.save(icon_path, optimize=True)

    for size, name in [(32, "favicon.png"), (180, "apple-touch-icon.png"), (512, "fifriz-icon-512.png")]:
        resized = icon_img.resize((size, size), Image.Resampling.LANCZOS)
        resized.save(os.path.join(OUT_DIR, name), optimize=True)

    print(f"Saved icon crop {x0},{y0},{x1},{y1} -> {icon_path}")


if __name__ == "__main__":
    main()
