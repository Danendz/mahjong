const SVG_NAME_MAP: Record<string, string> = {
  '1m': 'Man1', '2m': 'Man2', '3m': 'Man3', '4m': 'Man4', '5m': 'Man5',
  '6m': 'Man6', '7m': 'Man7', '8m': 'Man8', '9m': 'Man9',
  '1p': 'Pin1', '2p': 'Pin2', '3p': 'Pin3', '4p': 'Pin4', '5p': 'Pin5',
  '6p': 'Pin6', '7p': 'Pin7', '8p': 'Pin8', '9p': 'Pin9',
  '1s': 'Sou1', '2s': 'Sou2', '3s': 'Sou3', '4s': 'Sou4', '5s': 'Sou5',
  '6s': 'Sou6', '7s': 'Sou7', '8s': 'Sou8', '9s': 'Sou9',
  'we': 'Ton', 'ws': 'Nan', 'ww': 'Shaa', 'wn': 'Pei',
  'dz': 'Chun', 'df': 'Hatsu', 'db': 'Haku',
}

export function tileSvgUrl(code: string): string {
  const name = SVG_NAME_MAP[code]
  return name ? `/tiles/${name}.svg` : '/tiles/Back.svg'
}

export function tileBackUrl(): string {
  return '/tiles/Back.svg'
}
