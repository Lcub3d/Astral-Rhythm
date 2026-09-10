"""Compose README artwork from native Windows UI exports. Never bundles fonts."""
from pathlib import Path
import math
from PIL import Image, ImageDraw, ImageFont, ImageFilter

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / 'docs/images'
ASSETS = ROOT / 'src/assets/icons'
INK, MUTED, JADE = '#e0eae7', '#93aaa7', '#9dcabb'


def font(size: int, *, chinese: bool = False, bold: bool = False):
    choices = ([r'C:/Windows/Fonts/msyhbd.ttc' if bold else r'C:/Windows/Fonts/msyh.ttc',
                '/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc'] if chinese else
               [r'C:/Windows/Fonts/seguisb.ttf' if bold else r'C:/Windows/Fonts/segoeui.ttf',
                '/usr/share/fonts/truetype/lato/Lato-Semibold.ttf' if bold else '/usr/share/fonts/truetype/lato/Lato-Regular.ttf'])
    for p in choices:
        if Path(p).is_file():
            return ImageFont.truetype(p, size)
    raise RuntimeError('Install a system CJK/Latin font to render documentation; no font files are distributed.')


def background(width: int, height: int) -> Image.Image:
    im = Image.new('RGBA', (width, height), '#0c171b')
    d = ImageDraw.Draw(im)
    for y in range(height):
        t = y / height
        d.line((0, y, width, y), fill=(int(12+6*t), int(23+11*t), int(27+10*t)))
    for layer, color in enumerate(['#14292d', '#193136', '#1d363a']):
        points = [(0, height)]
        for x in range(0, width+10, 10):
            y = height*(.76+layer*.1) + math.sin(x/300+layer)*height*.075 + math.cos(x/155+layer)*height*.025
            points.append((x, y))
        points += [(width, height)]
        d.polygon(points, fill=color)
    d.rounded_rectangle((1, 1, width-2, height-2), radius=28, outline='#304347', width=2)
    return im


def text(im, xy, value, size, color=INK, *, chinese=False, bold=False):
    ImageDraw.Draw(im).text(xy, value, font=font(size, chinese=chinese, bold=bold), fill=color)


def paste(im, path: Path, xy, width: int):
    with Image.open(path) as src:
        s = src.convert('RGBA')
    s = s.resize((width, round(s.height*width/s.width)), Image.Resampling.LANCZOS)
    im.alpha_composite(s, xy)
    return s.size


def hero():
    im = background(1440, 970)
    d = ImageDraw.Draw(im)
    d.line((78, 86, 125, 86), fill=JADE, width=3)
    text(im, (142, 68), 'A QUIETER WAY TO WORK', 21, JADE)
    text(im, (78, 182), 'Astral', 100, bold=True)
    text(im, (78, 285), 'Rhythm', 100, bold=True)
    text(im, (82, 427), 'Work in rhythm.', 34, JADE)
    text(im, (82, 503), '七曜工作法', 34, chinese=True)
    text(im, (82, 563), '一周有章法，一日有节律。', 25, MUTED, chinese=True)
    text(im, (82, 608), 'Seven day stars. Twelve moments to refocus.', 21, MUTED)
    for j in range(7):
        paste(im, ASSETS/f'day-{(j+1)%7}-dark.png', (82+66*j, 703), 39)
    d.line((82, 792, 626, 792), fill='#3d5355', width=1)
    text(im, (82, 820), 'Windows  /  Offline  /  中文 + English', 20, MUTED, chinese=True)
    text(im, (82, 902), 'ASTRAL RHYTHM  ·  1.0', 16, '#74938c')
    shadow = Image.new('RGBA', im.size)
    ImageDraw.Draw(shadow).rounded_rectangle((798, 91, 1338, 758), radius=30, fill=(0, 0, 0, 140))
    im.alpha_composite(shadow.filter(ImageFilter.GaussianBlur(20)))
    paste(im, OUT/'card-zh-CN-dark.png', (790, 82), 550)
    d = ImageDraw.Draw(im)
    d.rounded_rectangle((791, 797, 1340, 887), radius=16, fill='#17282d', outline='#334b4d')
    paste(im, OUT/'taskbar-zh-CN-dark.png', (816, 808), 106)
    text(im, (974, 813), 'Always a glance away.', 20, INK)
    text(im, (974, 848), '透明任务栏 · 鼠标悬停展开', 16, MUTED, chinese=True)
    text(im, (793, 916), 'NATIVE WINDOWS RENDER  /  200% DPI', 16, '#8aa09b')
    im.convert('RGB').save(OUT/'hero.png', optimize=True)


def identities():
    im = background(1440, 620)
    text(im, (64, 39), 'Seven stars. Twelve shichen.', 34, bold=True)
    text(im, (66, 96), '七曜与十二时辰 · 各有形色，各司其时', 21, MUTED, chinese=True)
    days = ['月曜', '火曜', '水曜', '木曜', '金曜', '土曜', '日曜']
    roles = ['启 · Plan', '攻 · Tackle', '智 · Think', '扩 · Grow', '合 · Connect', '整 · Restore', '养 · Renew']
    for j, (label, role) in enumerate(zip(days, roles)):
        x = 64 + j*190
        paste(im, ASSETS/f'day-{(j+1)%7}-dark.png', (x+50, 166), 48)
        text(im, (x+45, 223), label, 25, chinese=True)
        text(im, (x+8, 271), role, 18, MUTED, chinese=True)
    ImageDraw.Draw(im).line((65, 330, 1373, 330), fill='#3a5354')
    for j, branch in enumerate('子丑寅卯辰巳午未申酉戌亥'):
        x = 70+j*109
        paste(im, ASSETS/f'hour-{j}-dark.png', (x+21, 381), 47)
        text(im, (x+31, 438), branch, 25, chinese=True)
        text(im, (x+2, 494), f'{(23+2*j)%24:02d}–{(1+2*j)%24:02d}', 19, MUTED)
    text(im, (66, 565), 'Original application artwork · Five-element accents · Independent day / hour identities', 19, MUTED)
    im.convert('RGB').save(OUT/'identities.png', optimize=True)


def main():
    required = ['card-zh-CN-dark.png', 'card-en-light.png', 'taskbar-zh-CN-dark.png', 'render-metadata.json']
    for name in required:
        if not (OUT/name).is_file():
            raise SystemExit(f'Missing native export: {name}. Run the Windows documentation test first.')
    hero()
    identities()
    print('README artwork composed from native exports; no desktop capture or bundled font files.')


if __name__ == '__main__':
    main()
