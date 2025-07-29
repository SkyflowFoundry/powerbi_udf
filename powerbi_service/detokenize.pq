let
    ToJsonText = (value as any) as text =>
        let
            jsonBinary = Json.FromValue(value), jsonText = Text.FromBinary(jsonBinary)
        in
            jsonText,
    // Detokenize multiple tokens in one API call
    Detokenize = (tokenList as list) as record =>
        let
            // Build the query string with multiple `token` query params
            queryParams = Text.Combine(List.Transform(tokenList, each "token=" & Text.From(_)), "&"),
            url = "https://APP_URL/api/detokenize?" & queryParams,
            headers = [#"Content-Type" = "application/json"],
            // Make API call and handle errors
            response = try Json.Document(Web.Contents(url, [Headers = headers])) otherwise {},
            // Normalize response into a token → value mapping
            tokenValueMap =
                if List.IsEmpty(response) then
                    []
                else
                    Record.FromList(List.Transform(response, each _[value]), List.Transform(response, each _[token]))
        in
            tokenValueMap,
    // Call API for each row and map detokenized values
    DetokenizeRecord = (inputRow as record) as record =>
        let
            // Extract token fields from row
            cardToken = Text.From(inputRow[card_number]),
            nameToken = Text.From(inputRow[name]),
            // List of tokens
            tokenList = {cardToken, nameToken},
            // Call detokenize API
            tokenMap = Detokenize(tokenList),
            // Add extracted values back to row
            rowWithCard = Record.AddField(
                inputRow, "Response_CardNumber", Record.FieldOrDefault(tokenMap, cardToken, null)
            ),
            rowWithName = Record.AddField(
                rowWithCard, "Response_Name", Record.FieldOrDefault(tokenMap, nameToken, null)
            )
        in
            rowWithName,
    AddApiResponses = (inputTable as table) as table =>
        Table.FromRecords(List.Transform(Table.ToRecords(inputTable), each DetokenizeRecord(_)))
in
    AddApiResponses
