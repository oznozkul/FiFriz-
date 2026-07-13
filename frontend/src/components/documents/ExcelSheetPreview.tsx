import { useMemo } from "react";
import { buildExcelMergeLayout, columnIndexToLetter, type ExcelSheetData } from "../../utils/xlsxPreview";

type ExcelSheetPreviewProps = {
  sheet: ExcelSheetData;
};

export function ExcelSheetPreview({ sheet }: ExcelSheetPreviewProps) {
  const columnLetters = useMemo(() => sheet.rows[0]?.map((_, index) => columnIndexToLetter(index)) ?? ["A"], [sheet.rows]);
  const mergeLayout = useMemo(() => buildExcelMergeLayout(sheet.merges), [sheet.merges]);

  return (
    <div className="excel-grid-scroll">
      <table className="excel-grid">
        <thead>
          <tr>
            <th className="excel-corner" scope="col" />
            {columnLetters.map((letter, index) => (
              <th key={`${letter}-${index}`} className="excel-col-header" scope="col" style={{ minWidth: sheet.columnWidths[index] ?? 88 }}>
                {letter}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {sheet.rows.map((row, rowIndex) => (
            <tr key={`row-${rowIndex}`}>
              <th className="excel-row-header" scope="row">
                {rowIndex + 1}
              </th>
              {row.map((cellValue, columnIndex) => {
                const cellKey = `${rowIndex},${columnIndex}`;
                if (mergeLayout.covered.has(cellKey)) return null;

                const merge = mergeLayout.mergeInfo.get(cellKey);

                return (
                  <td
                    key={`cell-${rowIndex}-${columnIndex}`}
                    className="excel-cell"
                    colSpan={merge?.colspan}
                    rowSpan={merge?.rowspan}
                    style={{ minWidth: sheet.columnWidths[columnIndex] ?? 88 }}
                  >
                    <span className="excel-cell-text">{cellValue}</span>
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
