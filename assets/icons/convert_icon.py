#!/usr/bin/env python3

import subprocess
import os
import sys
import struct
import zlib

OUTPUT_ICON = "AppDir/usr/share/icons/hicolor/48x48/apps/com.screenshooter.mcp.png"

def create_placeholder():
    def png_chunk(t, data):
        return struct.pack('>I', len(data)) + t + data + struct.pack('>I', zlib.crc32(t + data) & 0xffffffff)

    w, h = 48, 48
    ihdr = struct.pack('>IIBBBBB', w, h, 8, 2, 0, 0, 0)
    raw = b''.join(b'\x80\x00\x80\x00' for _ in range(w * h))
    idat = zlib.compress(raw, 9)
    png = b'\x89PNG\r\n\x1a\n' + png_chunk(b'IHDR', ihdr) + png_chunk(b'IDAT', idat) + png_chunk(b'IEND', b'')

    with open(OUTPUT_ICON, 'wb') as f:
        f.write(png)
    print(f"Created placeholder icon at {OUTPUT_ICON}")

def main():
    svg_path = "assets/icons/icon.svg"

    if not os.path.exists(svg_path):
        print(f"SVG file not found: {svg_path}")
        create_placeholder()
        return 1

    # Try rsvg-convert first
    if subprocess.call(["which", "rsvg-convert"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL) == 0:
        result = subprocess.run(
            ["rsvg-convert", "-w", "48", "-h", "48", "-o", OUTPUT_ICON, svg_path],
            capture_output=True
        )
        if result.returncode == 0:
            print(f"Converted icon using rsvg-convert")
            return 0

    # Try ImageMagick convert
    if subprocess.call(["which", "convert"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL) == 0:
        result = subprocess.run(
            ["convert", "-background", "none", "-resize", "48x48", svg_path, OUTPUT_ICON],
            capture_output=True
        )
        if result.returncode == 0:
            print(f"Converted icon using ImageMagick")
            return 0

    # Try cairosvg Python library
    try:
        import cairosvg
        cairosvg.svg2png(url=svg_path, write_to=OUTPUT_ICON, output_width=48, output_height=48)
        print(f"Converted icon using cairosvg")
        return 0
    except ImportError:
        pass

    print("No SVG converter available, creating placeholder")
    create_placeholder()
    return 0

if __name__ == "__main__":
    sys.exit(main())