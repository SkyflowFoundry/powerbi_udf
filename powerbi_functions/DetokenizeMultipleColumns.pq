(sourceTable as table, columnNames as list) as table =>
let
    // Process each column sequentially
    Result = List.Accumulate(
        columnNames,
        sourceTable,
        (currentTable, columnName) => DetokenizeColumn(currentTable, columnName)
    )
in
    Result