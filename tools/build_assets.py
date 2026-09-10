from pathlib import Path
import json, math
import cairosvg
from PIL import Image

ROOT=Path(__file__).resolve().parents[1]
A=ROOT/'src'/'assets'; (A/'icons').mkdir(parents=True,exist_ok=True); (A/'backgrounds').mkdir(exist_ok=True)
# Icons share a 32-unit grid and a round 1.55-unit stroke. No font glyphs or emoji.
DAYS=[
 ('日曜','日光','B88432','F2C771', '<circle cx="16" cy="16" r="5.2"/>' + ''.join(f'<path d="M{16+8*math.cos(i*math.pi/4):.2f},{16+8*math.sin(i*math.pi/4):.2f} L{16+11*math.cos(i*math.pi/4):.2f},{16+11*math.sin(i*math.pi/4):.2f}"/>' for i in range(8))),
 ('月曜','月白','7A74B6','B8B6EC','<path d="M23.5 20.5A10.2 10.2 0 0 1 11.5 6a10.2 10.2 0 1 0 12 14.5Z"/>'),
 ('火曜','朱砂','B95650','F3A59A','<circle cx="13" cy="19" r="7"/><path d="M18 14 26 6M19 6h7v7"/>'),
 ('水曜','澄水','357BA1','92C6E4','<circle cx="16" cy="14" r="6.1"/><path d="M10 4c0 5.3 12 5.3 12 0M16 20v9M12 25h8"/>'),
 ('木曜','青玉','3D846B','9BD5BA','<path d="M11 7c9-3 10 9-3 14h17M21 5v22M9 13c-1-3 0-5 2-6"/>'),
 ('金曜','香槟','9F7F34','E7CE8F','<circle cx="16" cy="12" r="7.2"/><path d="M16 19v10M11.5 25h9"/>'),
 ('土曜','岩褐','927358','D9B99A','<path d="M12 4v20M7 10h11M12 17c10-7 13 3 7 8-2 2-1 4 2 3"/>')
]
# Each Earthly Branch gets a different silhouette as well as a different color.
HOURS=[
 ('子时','夜航','596FAC','A8BAED','172741','385674','<path d="M10 10c-6 3-4 11 3 13 5 2 11 0 11-5 0-4-3-7-7-7M10 10C5 6 8 2 11 5c2-4 7-1 4 4M16 23c9 4 14 0 10-4M9 22l-3 3M12 24v2"/><circle cx="10.5" cy="13" r=".7" fill="currentColor" stroke="none"/>'),
 ('丑时','静夜','75649C','B9AADE','201F39','4A425F','<path d="M9 11C4 9 3 7 5 4c0 4 4 3 7 5M23 11c5-2 6-4 4-7 0 4-4 3-7 5M9 10c3-2 11-2 14 0l-2 13c-1 5-9 5-10 0Z M9 12 4 13l4 3M23 12l5 1-4 3M12 22h8"/><path d="M12 15h1M19 15h1M14 24h.2M18 24h.2"/>'),
 ('寅时','将明','8472A0','C8B5DF','2A2743','877388','<path d="M10 9C3 3 1 12 7 13M22 9c7-6 9 3 3 4M8 10c5-4 11-4 16 0 2 6 3 10-2 15-4 3-8 3-12 0-5-5-4-9-2-15Z M16 7v5M11 8l2 5M21 8l-2 5M8 15l3 2M24 15l-3 2M7 20l4 1M25 20l-4 1M13 21l3 2 3-2M16 23v3M11 17h1M20 17h1"/>'),
 ('卯时','晨曦','B97983','EFB5BA','E6CED2','F6ECE0','<path d="M13 15C5 0 11-1 16 12c1-17 8-14 4 1M13 15c-8 2-8 11 0 12h10c4-1 4-4 1-6 0-7-7-11-11-6ZM12 26c-1-6 6-8 6-2M23 25c5 0 5-5 1-5"/><circle cx="10" cy="18" r=".7" fill="currentColor" stroke="none"/>'),
 ('辰时','朝光','498C88','9ED5CB','DDECE5','EAF5EA','<path d="m13 6 4-3 6 4 5 2-5 4-8-1c-7 1-9 7-3 8 3 1 11-1 12 3 1 5-8 7-14 4-7-3-7-11 0-15M18 6l1-3M11 13l-4-4M9 21l-3 3M20 23l3 4M21 8h1"/>'),
 ('巳时','澄明','3F91A3','97D7E2','D5EAF0','F1F6E6','<path d="M9 7c6-8 19-2 12 5l-10 9c-5 5 0 8 9 5 6-3 3-8-2-5l-6 3M9 7l-4 2 5 3 5-2M11 8h.2M5 10l-2 2"/>'),
 ('午时','日中','BB7543','F3C48A','F0DDBB','F7EDDB','<path d="m13 11-3-5 7 1 5 6-3 3-6-1-1 5c7-2 12 2 12 7M13 9 9 12l-3 6 6 2M17 8l3-5v8M10 21l-4 6M15 21l-1 7M20 24l2 4M25 25c4-7-1-9-3-7M16 11h.2"/>'),
 ('未时','午后','698763','C3D5A3','E0E8CB','F3F1DE','<path d="M9 13c-9-1-7-13 0-8 4 3 2 6-1 4M23 13c9-1 7-13 0-8-4 3-2 6 1 4M10 12c3-3 9-3 12 0l-3 12-3 4-3-4ZM13 17h.2M19 17h.2M15 23h2M10 14l-5 1 4 3M22 14l5 1-4 3"/>'),
 ('申时','斜阳','9A8B45','DDCEA0','EADBBE','F4EEDD','<path d="M8 12c-8-3-8 9 0 8M24 12c8-3 8 9 0 8M8 11C9 1 23 1 24 11v10c-1 8-15 8-16 0ZM10 14c0-4 6-4 6 0 0-4 6-4 6 0M11 16h.2M21 16h.2M12 23c3 3 5 3 8 0M16 18v2"/>'),
 ('酉时','落霞','B56C63','EAB1A0','E1B7AB','EAE2CD','<path d="M12 9c-2-6 0-8 2-5 0-4 4-3 4 1 4-2 5 1 2 4M13 10l-1 7c-1 7 9 7 12 1l4-8c-6-2-9 0-9 6M12 15 6 13l6-3M12 18c-6 3-4 8 2 8h7M17 26v3M22 24v5M14 29h12M15 12h.2"/>'),
 ('戌时','灯晚','826A94','C6ADCF','55425F','B8979B','<path d="m10 9-5-6 1 13 4-2M22 9l5-6-1 13-4-2M10 10c4-2 8-2 12 0v10c-1 8-11 8-12 0ZM12 15h.2M20 15h.2M14 20h4l-2 2ZM16 22v3M8 23l-3 5M24 23l3 5"/>'),
 ('亥时','入梦','5B809E','A8C9E3','243548','617F94','<path d="M9 12 6 6l9 3M22 11l4-5 1 10M9 11c-5 2-6 11 1 15 6 3 14 0 14-7 0-6-7-10-15-8Z M23 18c7-3 7 4 4 3M10 26v3M20 26v3"/><rect x="9" y="17" width="10" height="6" rx="3"/><path d="M12 20h.2M16 20h.2M10 14h.2M19 14h.2"/>')
]


# V3: preserve the original line-art silhouettes, but give every label/icon
# a color in its own element family. Water becomes readable indigo on dark UI.
DAY_COLORS=[('97613C','F1B482'),('4A6483','ADBFD9'),('9E5142','EF9A82'),('396789','96BCD9'),('367359','A0D3B8'),('536671','D8E1E5'),('81682F','D8BD86')]
HOUR_COLORS=[('426080','97B3D8'),('7A673E','CEBC90'),('38715A','95C7AD'),('3B745B','A2D4B7'),('7D6834','D9C58F'),('9B5147','ECA08E'),('A25A45','F1AB8A'),('826C37','DDC68E'),('566975','CCD9E1'),('566973','DEE4E6'),('826638','D0B381'),('456783','A0BFDA')]
ELEMENTS_D=['火','水','火','水','木','金','土']
ELEMENTS_H=['水','土','木','木','土','火','火','土','金','金','土','水']
DAY_BASES=['192126','19212A','1B2227','172129','182528','1C232A','202529']
DAY_TOPS=['343029','2A303C','352D2A','263640','273A32','313A43','37332B']
HOUR_TOPS=['202C38','292C34','213337','263B38','2B3B38','29373B','303935','253B3D','303A3C','2E333B','2C2C37','22333C']

def hexmix(a,b,t):
 return ''.join(f'{round(int(a[i:i+2],16)*(1-t)+int(b[i:i+2],16)*t):02X}' for i in (0,2,4))
def lum(c):
 a=[int(c[i:i+2],16)/255 for i in (0,2,4)];a=[v/12.92 if v<=.04045 else ((v+.055)/1.055)**2.4 for v in a]
 return a[0]*.2126+a[1]*.7152+a[2]*.0722
def readable(c):
 while (lum('EDF1EF')+.05)/(lum(c)+.05)<4.6: c=hexmix(c,'000000',.03)
 return c

def icon_svg(shape,col):
 return f'<svg xmlns="http://www.w3.org/2000/svg" width="192" height="192" viewBox="0 0 32 32"><g fill="none" stroke="#{col}" color="#{col}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">{shape}</g></svg>'
def output(folder,name,svg,w,h):
 base=A/folder/name
 base.with_suffix('.svg').write_text(svg,encoding='utf-8')
 cairosvg.svg2png(bytestring=svg.encode(),write_to=str(base.with_suffix('.png')),output_width=w,output_height=h)

def background(kind,index,col,base,top,light):
 # Deliberately no lettering in artwork. The app paints real native text.
 opacity='.055' if light else '.20'
 shift=index*11
 moon=f'<circle cx="760" cy="57" r="23" fill="#{col}" opacity=".18"/>'
 if kind=='day' and index==1 or kind=='hour' and index in (0,1,2,10,11):
  moon=f'<path d="M778 44a24 24 0 1 1-31-23 23 23 0 0 0 31 23Z" fill="#{col}" opacity=".16"/>'
 if kind=='hour':
  ridges=f'<path d="M0 365Q220 300 360 255Q450 {145+shift%18} 520 194T718 169T960 143V500H0Z" fill="#{hexmix(col,base,.9)}" opacity=".7"/><path d="M0 500Q350 400 540 304Q630 181 724 238T960 235V500H0Z" fill="#{base}" opacity=".64"/>'
 else:
  ridges=f'<g stroke="#{col}" stroke-opacity="{opacity}" stroke-width="1.2" fill="none"><path d="M655 206Q710 193 748 126T834 151T961 174"/><path d="M665 229Q722 217 761 157T846 177T965 203"/><path d="M683 246Q731 237 772 190T852 204T966 231"/></g>'
 return f'''<svg xmlns="http://www.w3.org/2000/svg" width="960" height="500" viewBox="0 0 960 500"><defs><linearGradient id="g" x1="0" y1="0" x2="0" y2="1"><stop stop-color="#{top}"/><stop offset=".73" stop-color="#{base}"/><stop offset="1" stop-color="#{base}"/></linearGradient><linearGradient id="veil"><stop stop-color="#{base}" stop-opacity=".28"/><stop offset=".48" stop-color="#{base}" stop-opacity=".24"/><stop offset="1" stop-color="#{base}" stop-opacity="0"/></linearGradient><linearGradient id="fade" x1="0" y1="0" x2="0" y2="1"><stop offset=".60" stop-color="#{base}" stop-opacity="0"/><stop offset=".98" stop-color="#{base}" stop-opacity="1"/><stop offset="1" stop-color="#{base}" stop-opacity="1"/></linearGradient></defs><rect width="960" height="500" fill="url(#g)"/>{moon}{ridges}<rect width="960" height="500" fill="url(#veil)"/><rect width="960" height="500" fill="url(#fade)"/></svg>'''

manifest={'name':'山岚·凝时','version':3,'days':[],'hours':[]}
for kind,items,colors,els in [('day',DAYS,DAY_COLORS,ELEMENTS_D),('hour',HOURS,HOUR_COLORS,ELEMENTS_H)]:
 for i,item in enumerate(items):
  name,title=item[:2];shape=item[-1];lc,dc=colors[i];lc=readable(lc)
  darkbase=DAY_BASES[i] if kind=='day' else hexmix(HOUR_TOPS[i],'172329',.90)
  lightbase=hexmix(lc,'F6F8F7',.975)
  record={'name':name,'title':title,'element':els[i],'light':lc,'dark':dc,'soft_light':lightbase,'soft_dark':darkbase}
  manifest['days' if kind=='day' else 'hours'].append(record)
  for mode,col,base in [('dark',dc,darkbase),('light',lc,lightbase)]:
   output('icons',f'{kind}-{i}-{mode}',icon_svg(shape,col),192,192)
   top=(DAY_TOPS[i] if kind=='day' else HOUR_TOPS[i]) if mode=='dark' else hexmix(col,base,.93)
   output('backgrounds',f'{kind}-{i}-{mode}',background(kind,i,col,base,top,mode=='light'),960,500)
(A/'design.json').write_text(json.dumps(manifest,ensure_ascii=False,indent=2),encoding='utf-8')
# Section icons match the small sun and earthly-branch constellation in the reference.
sun=DAYS[0][-1]
constellation='<path d="M8 13 16 6l8 7-3 10H11ZM16 6v11M8 13l8 4 8-4M11 23l5-6 5 6M11 23l-1 5M21 23l1 5"/><circle cx="16" cy="5" r="2.4"/><circle cx="7" cy="13" r="2.2"/><circle cx="25" cy="13" r="2.2"/>'
for mode,suncol,hcol in [('dark','EAB07D','9BCDBD'),('light','93613F','477667')]:
 output('icons','header-day-'+mode,icon_svg(sun,suncol),192,192)
 output('icons','header-hour-'+mode,icon_svg(constellation,hcol),192,192)
# Keep a restrained Windows application icon. No font file is distributed.
appsvg='<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 64 64"><defs><linearGradient id="b" x2="1" y2="1"><stop stop-color="#223943"/><stop offset="1" stop-color="#3C5858"/></linearGradient></defs><rect width="64" height="64" rx="16" fill="url(#b)"/><g fill="none" stroke-linecap="round"><path d="M17 32a15 15 0 0 1 30 0M11 34h42" stroke="#DFBE8E" stroke-width="2.3"/><path d="M18 42h28M24 49h16" stroke="#AAD0BD" stroke-width="2.3"/><path d="M32 10v5M13 19l3 3M51 19l-3 3" stroke="#DFBE8E" stroke-width="2"/></g></svg>'
output('icons','app',appsvg,256,256)
Image.open(A/'icons'/'app.png').save(ROOT/'AstralRhythm.ico',sizes=[(16,16),(24,24),(32,32),(48,48),(64,64),(128,128),(256,256)])
print('V3 assets:',len(list(A.rglob('*.png'))),'PNG + editable SVG masters')
