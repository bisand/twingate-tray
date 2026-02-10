#!/usr/bin/env python3
"""Convert Font Awesome SVG path data to flattened polygon points for Go embedding.

Parses SVG path d="" attributes, flattens cubic bezier curves into line segments,
and outputs normalized [0,1] coordinate arrays suitable for Go source code.
"""

import math
import re
import sys


def parse_svg_path(d):
    """Parse SVG path d attribute into a list of (command, args) tuples."""
    # Tokenize: split into commands and numbers
    tokens = re.findall(r'[MmLlHhVvCcSsQqTtAaZz]|[-+]?(?:\d+\.?\d*|\.\d+)(?:[eE][-+]?\d+)?', d)
    
    commands = []
    i = 0
    while i < len(tokens):
        if tokens[i].isalpha():
            cmd = tokens[i]
            i += 1
            args = []
            while i < len(tokens) and not tokens[i].isalpha():
                args.append(float(tokens[i]))
                i += 1
            commands.append((cmd, args))
        else:
            i += 1
    
    return commands


def cubic_bezier(p0, p1, p2, p3, steps=16):
    """Flatten a cubic bezier curve into line segments."""
    points = []
    for i in range(1, steps + 1):
        t = i / steps
        t2 = t * t
        t3 = t2 * t
        mt = 1 - t
        mt2 = mt * mt
        mt3 = mt2 * mt
        
        x = mt3 * p0[0] + 3 * mt2 * t * p1[0] + 3 * mt * t2 * p2[0] + t3 * p3[0]
        y = mt3 * p0[1] + 3 * mt2 * t * p1[1] + 3 * mt * t2 * p2[1] + t3 * p3[1]
        points.append((x, y))
    
    return points


def flatten_path(commands):
    """Convert parsed SVG path commands into a list of polygon outlines (list of point lists)."""
    polygons = []
    current_polygon = []
    cx, cy = 0, 0  # Current point
    sx, sy = 0, 0  # Start of current subpath
    last_cp = None  # Last control point for S command
    
    for cmd, args in commands:
        if cmd == 'M':
            if current_polygon:
                polygons.append(current_polygon)
            cx, cy = args[0], args[1]
            sx, sy = cx, cy
            current_polygon = [(cx, cy)]
            # Implicit L commands for remaining pairs
            i = 2
            while i + 1 < len(args):
                cx, cy = args[i], args[i+1]
                current_polygon.append((cx, cy))
                i += 2
            last_cp = None
        
        elif cmd == 'm':
            if current_polygon:
                polygons.append(current_polygon)
            cx += args[0]
            cy += args[1]
            sx, sy = cx, cy
            current_polygon = [(cx, cy)]
            i = 2
            while i + 1 < len(args):
                cx += args[i]
                cy += args[i+1]
                current_polygon.append((cx, cy))
                i += 2
            last_cp = None
        
        elif cmd == 'L':
            i = 0
            while i + 1 < len(args):
                cx, cy = args[i], args[i+1]
                current_polygon.append((cx, cy))
                i += 2
            last_cp = None
        
        elif cmd == 'l':
            i = 0
            while i + 1 < len(args):
                cx += args[i]
                cy += args[i+1]
                current_polygon.append((cx, cy))
                i += 2
            last_cp = None
        
        elif cmd == 'H':
            for v in args:
                cx = v
                current_polygon.append((cx, cy))
            last_cp = None
        
        elif cmd == 'h':
            for v in args:
                cx += v
                current_polygon.append((cx, cy))
            last_cp = None
        
        elif cmd == 'V':
            for v in args:
                cy = v
                current_polygon.append((cx, cy))
            last_cp = None
        
        elif cmd == 'v':
            for v in args:
                cy += v
                current_polygon.append((cx, cy))
            last_cp = None
        
        elif cmd == 'C':
            i = 0
            while i + 5 < len(args):
                p0 = (cx, cy)
                p1 = (args[i], args[i+1])
                p2 = (args[i+2], args[i+3])
                p3 = (args[i+4], args[i+5])
                pts = cubic_bezier(p0, p1, p2, p3)
                current_polygon.extend(pts)
                cx, cy = p3
                last_cp = p2
                i += 6
        
        elif cmd == 'c':
            i = 0
            while i + 5 < len(args):
                p0 = (cx, cy)
                p1 = (cx + args[i], cy + args[i+1])
                p2 = (cx + args[i+2], cy + args[i+3])
                p3 = (cx + args[i+4], cy + args[i+5])
                pts = cubic_bezier(p0, p1, p2, p3)
                current_polygon.extend(pts)
                last_cp = p2
                cx, cy = p3
                i += 6
        
        elif cmd == 'S':
            i = 0
            while i + 3 < len(args):
                p0 = (cx, cy)
                if last_cp:
                    p1 = (2 * cx - last_cp[0], 2 * cy - last_cp[1])
                else:
                    p1 = (cx, cy)
                p2 = (args[i], args[i+1])
                p3 = (args[i+2], args[i+3])
                pts = cubic_bezier(p0, p1, p2, p3)
                current_polygon.extend(pts)
                last_cp = p2
                cx, cy = p3
                i += 4
        
        elif cmd == 's':
            i = 0
            while i + 3 < len(args):
                p0 = (cx, cy)
                if last_cp:
                    p1 = (2 * cx - last_cp[0], 2 * cy - last_cp[1])
                else:
                    p1 = (cx, cy)
                p2 = (cx + args[i], cy + args[i+1])
                p3 = (cx + args[i+2], cy + args[i+3])
                pts = cubic_bezier(p0, p1, p2, p3)
                current_polygon.extend(pts)
                last_cp = p2
                cx, cy = p3
                i += 4
        
        elif cmd in ('Z', 'z'):
            cx, cy = sx, sy
            if current_polygon:
                current_polygon.append((sx, sy))
                polygons.append(current_polygon)
                current_polygon = []
            last_cp = None
    
    if current_polygon:
        polygons.append(current_polygon)
    
    return polygons


def normalize_polygons(polygons, vb_width, vb_height):
    """Normalize polygon coordinates to [0, 1] range based on viewBox."""
    result = []
    for poly in polygons:
        normalized = []
        for x, y in poly:
            normalized.append((x / vb_width, y / vb_height))
        result.append(normalized)
    return result


def format_go_array(polygons, var_name):
    """Format polygon data as Go source code."""
    lines = []
    lines.append(f'// {var_name} contains the Font Awesome icon path as normalized [0,1] polygon outlines.')
    lines.append(f'// Each sub-slice is a closed polygon outline. Uses even-odd fill rule.')
    lines.append(f'var {var_name} = [][][2]float64{{')
    
    for i, poly in enumerate(polygons):
        lines.append(f'\t// Polygon {i} ({len(poly)} points)')
        lines.append('\t{')
        for j, (x, y) in enumerate(poly):
            comma = ',' if j < len(poly) - 1 else ','
            lines.append(f'\t\t{{{x:.6f}, {y:.6f}}}{comma}')
        lines.append('\t},')
    
    lines.append('}')
    return '\n'.join(lines)


# Font Awesome lock (solid) - viewBox 0 0 448 512
lock_path = "M144 144l0 48 160 0 0-48c0-44.2-35.8-80-80-80s-80 35.8-80 80zM80 192l0-48C80 64.5 144.5 0 224 0s144 64.5 144 144l0 48 16 0c35.3 0 64 28.7 64 64l0 192c0 35.3-28.7 64-64 64L64 512c-35.3 0-64-28.7-64-64L0 256c0-35.3 28.7-64 64-64l16 0z"

# Font Awesome unlock (solid) - viewBox 0 0 448 512
unlock_path = "M144 144c0-44.2 35.8-80 80-80c31.9 0 59.4 18.6 72.3 45.7c7.6 16 26.7 22.8 42.6 15.2s22.8-26.7 15.2-42.6C331 33.7 281.5 0 224 0C144.5 0 80 64.5 80 144l0 48-16 0c-35.3 0-64 28.7-64 64L0 448c0 35.3 28.7 64 64 64l320 0c35.3 0 64-28.7 64-64l0-192c0-35.3-28.7-64-64-64l-240 0 0-48z"


lock_commands = parse_svg_path(lock_path)
lock_polygons = flatten_path(lock_commands)
lock_normalized = normalize_polygons(lock_polygons, 448, 512)

unlock_commands = parse_svg_path(unlock_path)
unlock_polygons = flatten_path(unlock_commands)
unlock_normalized = normalize_polygons(unlock_polygons, 448, 512)

print(f"// Lock icon: {len(lock_polygons)} polygons, total {sum(len(p) for p in lock_polygons)} points")
for i, p in enumerate(lock_polygons):
    print(f"//   Polygon {i}: {len(p)} points")
print()
print(f"// Unlock icon: {len(unlock_polygons)} polygons, total {sum(len(p) for p in unlock_polygons)} points")
for i, p in enumerate(unlock_polygons):
    print(f"//   Polygon {i}: {len(p)} points")
print()

print(format_go_array(lock_normalized, "faLockPolygons"))
print()
print(format_go_array(unlock_normalized, "faUnlockPolygons"))
