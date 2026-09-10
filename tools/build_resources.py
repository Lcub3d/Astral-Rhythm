from pathlib import Path
import struct, re
R=Path(__file__).resolve().parents[1]
source=(R/'src'/'locale.go').read_text(encoding='utf-8')
VERSION=re.search(r'const appVersion = "([0-9]+\.[0-9]+\.[0-9]+)"', source)[1]
ENGLISH=re.search(r'const appNameEN = "([^"]+)"', source)[1]
MAJOR,MINOR,PATCH=map(int,VERSION.split('.'))
def align(b): return b+b'\0'*((-len(b))%4)
def block(key,value=b'',children=(),typ=1,value_len=None):
 s=align(b'\0'*6+(key+'\0').encode('utf-16le'))+value
 for c in children:s=align(s)+c
 if value_len is None:value_len=len(value)//2 if typ==1 else len(value)
 return struct.pack('<HHH',len(s),value_len,typ)+s[6:]
def string_table(language, product):
 strings=[]
 for k,v in [('CompanyName',ENGLISH),('FileDescription',product),('FileVersion',VERSION),('InternalName','AstralRhythm'),('OriginalFilename','AstralRhythm.exe'),('ProductName',product),('ProductVersion',VERSION)]:
  strings.append(block(k,(v+'\0').encode('utf-16le')))
 return block(language,children=strings)
fixed=struct.pack('<13I',0xFEEF04BD,0x10000,(MAJOR<<16)|MINOR,PATCH<<16,(MAJOR<<16)|MINOR,PATCH<<16,0x3F,0,0x40004,1,0,0,0)
ver=block('VS_VERSION_INFO',fixed,[block('StringFileInfo',children=[string_table('080404B0','七曜工作法'),string_table('040904B0',ENGLISH)]),block('VarFileInfo',children=[block('Translation',struct.pack('<HHHH',0x804,1200,0x409,1200),typ=0)])],typ=0)
manifest=b'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?><assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0"><assemblyIdentity version="VERSION.0" processorArchitecture="amd64" name="Astral.Desktop" type="win32"/><description>Astral Rhythm</description><trustInfo xmlns="urn:schemas-microsoft-com:asm.v3"><security><requestedPrivileges><requestedExecutionLevel level="asInvoker" uiAccess="false"/></requestedPrivileges></security></trustInfo><compatibility xmlns="urn:schemas-microsoft-com:compatibility.v1"><application><supportedOS Id="{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}"/></application></compatibility><application xmlns="urn:schemas-microsoft-com:asm.v3"><windowsSettings><dpiAwareness xmlns="http://schemas.microsoft.com/SMI/2016/WindowsSettings">PerMonitorV2, PerMonitor</dpiAwareness><dpiAware xmlns="http://schemas.microsoft.com/SMI/2005/WindowsSettings">true/pm</dpiAware><longPathAware xmlns="http://schemas.microsoft.com/SMI/2016/WindowsSettings">true</longPathAware></windowsSettings></application></assembly>'''
manifest=manifest.replace(b'VERSION',VERSION.encode('ascii'))
ico=(R/'AstralRhythm.ico').read_bytes();count=struct.unpack_from('<H',ico,4)[0];icons={};group=struct.pack('<HHH',0,1,count)
for i in range(count):
 w,h,c,z,planes,bpp,n,off=struct.unpack_from('<BBBBHHII',ico,6+16*i);icons[i+1]=ico[off:off+n];group+=struct.pack('<BBBBHHIH',w,h,c,z,planes,bpp,n,i+1)
resources={3:{i:{0x804:v} for i,v in icons.items()},14:{1:{0x804:group}},16:{1:{0x804:ver}},24:{1:{0x804:manifest}}}
raw=bytearray();rel=[]
def alloc(n):
 while len(raw)%4:raw.append(0)
 off=len(raw);raw.extend(b'\0'*n);return off
def tree(d):
 off=alloc(16+8*len(d));struct.pack_into('<IIHHHH',raw,off,0,0,0,0,0,len(d))
 for j,(ident,val) in enumerate(sorted(d.items())):
  if isinstance(val,dict):child=tree(val)|0x80000000
  else:
   child=alloc(16)
   while len(raw)%4:raw.append(0)
   payload=alloc(len(val));raw[payload:payload+len(val)]=val
   struct.pack_into('<IIII',raw,child,payload,len(val),0,0);rel.append(child)
  struct.pack_into('<II',raw,off+16+j*8,ident,child)
 return off
tree(resources)
raw=bytearray(align(bytes(raw)))
reloc=b''.join(struct.pack('<IIH',off,0,3) for off in rel)
ptrsyms=60+len(raw)+len(reloc)
header=struct.pack('<HHIIIHH',0x8664,1,0,ptrsyms,1,0,0)
section=struct.pack('<8sIIIIIIHHI',b'.rsrc\0\0\0',0,0,len(raw),60,60+len(raw),0,len(rel),0,0x40000040)
sym=struct.pack('<8sIhHBB',b'.rsrc\0\0\0',0,1,0,3,0)
(R/'src'/'resources_windows_amd64.syso').write_bytes(header+section+raw+reloc+sym+struct.pack('<I',4))
print('COFF resources written:',count,'icon sizes; version',VERSION,'; PerMonitorV2 manifest')
