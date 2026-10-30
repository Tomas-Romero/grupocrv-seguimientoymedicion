"""Calcula el contraste WCAG de los pares de colores del diseno (spec T-024, RN-T024-4) a partir del CSS compilado.

Uso: python scripts/contraste.py [web/static/css/app.css]

Lee las variables --color-* de Tailwind (estan en oklch), las pasa a sRGB y calcula la relacion de contraste de cada par. Sale con 1 si algun par
de texto queda debajo de 4,5:1 o algun borde de campo o boton queda debajo de 3:1. No usa librerias externas.
"""
import math
import re
import sys

ruta = sys.argv[1] if len(sys.argv) > 1 else "web/static/css/app.css"
css = open(ruta, encoding="utf-8").read()


def oklch_a_srgb(l, c, h):
    a, b = c * math.cos(math.radians(h)), c * math.sin(math.radians(h))
    l_ = (l + 0.3963377774 * a + 0.2158037573 * b) ** 3
    m_ = (l - 0.1055613458 * a - 0.0638541728 * b) ** 3
    s_ = (l - 0.0894841775 * a - 1.2914855480 * b) ** 3
    lin = (
        4.0767416621 * l_ - 3.3077115913 * m_ + 0.2309699292 * s_,
        -1.2684380046 * l_ + 2.6097574011 * m_ - 0.3413193965 * s_,
        -0.0041960863 * l_ - 0.7034186147 * m_ + 1.7076147010 * s_,
    )
    return tuple(min(1.0, max(0.0, v)) for v in lin)  # lineal, recortado al gamut sRGB


colores = {"white": (1.0, 1.0, 1.0)}
for nombre, l, c, h in re.findall(r"--color-([a-z]+-\d+):oklch\(([\d.]+%?) ([\d.]+) ([\d.]+)\)", css):
    valor = float(l[:-1]) / 100 if l.endswith("%") else float(l)
    colores[nombre] = oklch_a_srgb(valor, float(c), float(h))


def luminancia(rgb_lineal):
    r, g, b = rgb_lineal
    return 0.2126 * r + 0.7152 * g + 0.0722 * b


def contraste(a, b):
    la, lb = luminancia(colores[a]), luminancia(colores[b])
    claro, oscuro = max(la, lb), min(la, lb)
    return (claro + 0.05) / (oscuro + 0.05)


# (que es, texto o borde, fondo, minimo)
pares = [
    ("insignia primaria", "indigo-900", "indigo-100", 4.5),
    ("insignia info", "sky-900", "sky-100", 4.5),
    ("insignia ok (completado)", "emerald-900", "emerald-100", 4.5),
    ("insignia atencion (pendiente)", "amber-900", "amber-100", 4.5),
    ("insignia apagada (wont)", "slate-600", "white", 4.5),
    ("insignia neutra (could)", "slate-800", "slate-100", 4.5),
    ("boton primario", "white", "indigo-700", 4.5),
    ("boton primario (hover)", "white", "indigo-800", 4.5),
    ("boton secundario", "slate-800", "white", 4.5),
    ("enlace sobre blanco", "indigo-700", "white", 4.5),
    ("enlace sobre gris claro", "indigo-700", "slate-50", 4.5),
    ("enlace (hover)", "indigo-900", "white", 4.5),
    ("menu activo", "indigo-800", "indigo-50", 4.5),
    ("menu inactivo sobre blanco", "slate-700", "white", 4.5),
    ("menu inactivo sobre la barra del proyecto", "slate-700", "slate-50", 4.5),
    ("menu (hover)", "slate-900", "slate-100", 4.5),
    ("chip del proyecto", "slate-600", "white", 4.5),
    ("encabezado de tabla", "slate-600", "slate-50", 4.5),
    ("numero de fila", "slate-600", "white", 4.5),
    ("cifra de una tarjeta", "slate-800", "slate-50", 4.5),
    ("titulo de la velocidad", "indigo-800", "white", 4.5),
    ("alerta", "red-900", "red-50", 4.5),
    ("campo: borde", "slate-500", "white", 3.0),
    ("boton secundario: borde", "slate-500", "white", 3.0),
    ("barra de Story Points planificados", "slate-500", "white", 3.0),
    ("barra de Story Points completados", "indigo-700", "white", 3.0),
]

fallos = 0
for nombre, a, b, minimo in pares:
    if a not in colores or b not in colores:
        print(f"FALTA  {nombre}: no encuentro {a} o {b} en el CSS")
        fallos += 1
        continue
    r = contraste(a, b)
    marca = "ok   " if r >= minimo else "FALLA"
    fallos += r < minimo
    print(f"{marca} {r:5.2f}:1 (minimo {minimo}:1)  {nombre}  [{a} sobre {b}]")
sys.exit(1 if fallos else 0)
