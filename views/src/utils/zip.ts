// Minimal ZIP writer (stored/uncompressed entries only) — no third-party deps.
// Produces a valid .zip containing plain-text files, sufficient for bundling
// converted JSON output for a single-click download.

const CRC_TABLE = buildCrcTable()

function buildCrcTable(): Uint32Array {
  const table = new Uint32Array(256)
  for (let n = 0; n < 256; n++) {
    let c = n
    for (let k = 0; k < 8; k++) {
      c = c & 1 ? (0xedb88320 ^ (c >>> 1)) : c >>> 1
    }
    table[n] = c >>> 0
  }
  return table
}

function crc32(data: Uint8Array): number {
  let crc = 0xffffffff
  for (let i = 0; i < data.length; i++) {
    crc = CRC_TABLE[(crc ^ data[i]) & 0xff] ^ (crc >>> 8)
  }
  return (crc ^ 0xffffffff) >>> 0
}

function dosDateTime(date: Date): { time: number; date: number } {
  const time =
    (date.getHours() << 11) | (date.getMinutes() << 5) | (Math.floor(date.getSeconds() / 2))
  const dosDate =
    (((date.getFullYear() - 1980) & 0x7f) << 9) | ((date.getMonth() + 1) << 5) | date.getDate()
  return { time, date: dosDate }
}

export interface ZipEntry {
  name: string
  content: string
}

export function buildZip(entries: ZipEntry[]): Blob {
  const encoder = new TextEncoder()
  const { time, date } = dosDateTime(new Date())
  const chunks: Uint8Array[] = []
  const centralRecords: Uint8Array[] = []
  let offset = 0

  for (const entry of entries) {
    const nameBytes = encoder.encode(entry.name)
    const dataBytes = encoder.encode(entry.content)
    const crc = crc32(dataBytes)

    const localHeader = new DataView(new ArrayBuffer(30))
    localHeader.setUint32(0, 0x04034b50, true)
    localHeader.setUint16(4, 20, true) // version needed
    localHeader.setUint16(6, 0x0800, true) // UTF-8 filenames
    localHeader.setUint16(8, 0, true) // stored, no compression
    localHeader.setUint16(10, time, true)
    localHeader.setUint16(12, date, true)
    localHeader.setUint32(14, crc, true)
    localHeader.setUint32(18, dataBytes.length, true)
    localHeader.setUint32(22, dataBytes.length, true)
    localHeader.setUint16(26, nameBytes.length, true)
    localHeader.setUint16(28, 0, true) // extra field length

    chunks.push(new Uint8Array(localHeader.buffer), nameBytes, dataBytes)

    const central = new DataView(new ArrayBuffer(46))
    central.setUint32(0, 0x02014b50, true)
    central.setUint16(4, 20, true) // version made by
    central.setUint16(6, 20, true) // version needed
    central.setUint16(8, 0x0800, true)
    central.setUint16(10, 0, true)
    central.setUint16(12, time, true)
    central.setUint16(14, date, true)
    central.setUint32(16, crc, true)
    central.setUint32(20, dataBytes.length, true)
    central.setUint32(24, dataBytes.length, true)
    central.setUint16(28, nameBytes.length, true)
    central.setUint16(30, 0, true) // extra length
    central.setUint16(32, 0, true) // comment length
    central.setUint16(34, 0, true) // disk number start
    central.setUint16(36, 0, true) // internal attributes
    central.setUint32(38, 0, true) // external attributes
    central.setUint32(42, offset, true) // offset of local header

    centralRecords.push(new Uint8Array(central.buffer), nameBytes)

    offset += localHeader.byteLength + nameBytes.length + dataBytes.length
  }

  const centralStart = offset
  let centralSize = 0
  for (const chunk of centralRecords) centralSize += chunk.length

  const end = new DataView(new ArrayBuffer(22))
  end.setUint32(0, 0x06054b50, true)
  end.setUint16(4, 0, true)
  end.setUint16(6, 0, true)
  end.setUint16(8, entries.length, true)
  end.setUint16(10, entries.length, true)
  end.setUint32(12, centralSize, true)
  end.setUint32(16, centralStart, true)
  end.setUint16(20, 0, true)

  return new Blob([...chunks, ...centralRecords, new Uint8Array(end.buffer)], {
    type: 'application/zip',
  })
}
