(sourceTable as table, columnName as text) as table =>
let
    // Extract tokens from the column
    TokenColumn = Table.Column(sourceTable, columnName),
    UniqueTokens = List.Distinct(List.Select(TokenColumn, each _ <> null and _ <> "")),
    
    // Create token objects JSON manually
    TokenObjectsJson = "{""tokens"": [" & 
        Text.Combine(
            List.Transform(UniqueTokens, each "{""token"": """ & _ & """}"),
            ", "
        ) & 
        "], ""redaction"": ""PLAIN_TEXT""}",
    
    // Make the API call
    Response = Web.Contents(
        "http://localhost:8080/detokenize",  //replace with your skyflow detokenizer url (Local or Lambda or Docker Url)
        [
            Headers = [#"Content-Type" = "application/json"],
            Content = Text.ToBinary(TokenObjectsJson),
            ManualStatusHandling = {200, 400, 404, 500}
        ]
    ),
    
    // Process response using the working approach
    Bytes = Binary.ToList(Response),
    CharCodes = List.Transform(Bytes, each Character.FromNumber(_)),
    ResponseText = Text.Combine(CharCodes),
    
    // Parse the JSON response
    ParsedResponse = Json.Document(ResponseText),
    
    // Convert to a table format
    ResponseTable = Record.ToTable(ParsedResponse),
    
    // Join back to the original table
    JoinedTable = Table.NestedJoin(
        sourceTable,
        {columnName},
        ResponseTable,
        {"Name"},
        "TokenValues",
        JoinKind.LeftOuter
    ),
    
    // Expand the joined column to add the detokenized values
    ResultTable = Table.ExpandTableColumn(
        JoinedTable,
        "TokenValues",
        {"Value"},
        {"Detokenized_" & columnName}
    )
in
    ResultTable
