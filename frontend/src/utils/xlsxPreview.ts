import * as XLSX from "xlsx";

export type ExcelSheetData = {
  name: string;
  rows: string[][];
  merges: XLSX.Range[];
  columnWidths: number[];
};

export function parseXlsxWorkbook(buffer: ArrayBuffer): ExcelSheetData[] {
  const workbook = XLSX.read(buffer, { type: "array", cellText: true, cellDates: true });

  return workbook.SheetNames.map((name) => parseExcelSheet(workbook.Sheets[name], name));
}

function parseExcelSheet(worksheet: XLSX.WorkSheet, name: string): ExcelSheetData {
  if (!worksheet["!ref"]) {
    return { name, rows: [[""]], merges: [], columnWidths: [88] };
  }

  const range = XLSX.utils.decode_range(worksheet["!ref"]);
  const rowOffset = range.s.r;
  const columnOffset = range.s.c;
  const rows: string[][] = [];

  for (let rowIndex = range.s.r; rowIndex <= range.e.r; rowIndex += 1) {
    const row: string[] = [];
    for (let columnIndex = range.s.c; columnIndex <= range.e.c; columnIndex += 1) {
      const address = XLSX.utils.encode_cell({ r: rowIndex, c: columnIndex });
      row.push(formatExcelCellValue(worksheet[address]));
    }
    rows.push(row);
  }

  const columnCount = rows[0]?.length ?? 1;
  const columnWidths = Array.from({ length: columnCount }, (_, index) => {
    const column = worksheet["!cols"]?.[index + columnOffset];
    if (column?.wpx) return Math.max(column.wpx, 48);
    if (column?.width) return Math.max(column.width * 7, 48);
    return 88;
  });

  const merges = (worksheet["!merges"] ?? []).map((merge) => ({
    s: { r: merge.s.r - rowOffset, c: merge.s.c - columnOffset },
    e: { r: merge.e.r - rowOffset, c: merge.e.c - columnOffset },
  }));

  return {
    name,
    rows,
    merges,
    columnWidths,
  };
}

function formatExcelCellValue(cell?: XLSX.CellObject): string {
  if (!cell) return "";
  if (cell.w) return cell.w;
  if (cell.v == null) return "";
  if (cell.t === "d" && cell.v instanceof Date) {
    return cell.v.toLocaleString("tr-TR");
  }
  return String(cell.v);
}

export function columnIndexToLetter(index: number): string {
  let value = index + 1;
  let letters = "";

  while (value > 0) {
    const remainder = (value - 1) % 26;
    letters = String.fromCharCode(65 + remainder) + letters;
    value = Math.floor((value - 1) / 26);
  }

  return letters;
}

export type ExcelMergeLayout = {
  covered: Set<string>;
  mergeInfo: Map<string, { colspan: number; rowspan: number }>;
};

export function buildExcelMergeLayout(merges: XLSX.Range[]): ExcelMergeLayout {
  const covered = new Set<string>();
  const mergeInfo = new Map<string, { colspan: number; rowspan: number }>();

  for (const merge of merges) {
    const key = `${merge.s.r},${merge.s.c}`;
    mergeInfo.set(key, {
      colspan: merge.e.c - merge.s.c + 1,
      rowspan: merge.e.r - merge.s.r + 1,
    });

    for (let rowIndex = merge.s.r; rowIndex <= merge.e.r; rowIndex += 1) {
      for (let columnIndex = merge.s.c; columnIndex <= merge.e.c; columnIndex += 1) {
        if (rowIndex === merge.s.r && columnIndex === merge.s.c) continue;
        covered.add(`${rowIndex},${columnIndex}`);
      }
    }
  }

  return { covered, mergeInfo };
}
