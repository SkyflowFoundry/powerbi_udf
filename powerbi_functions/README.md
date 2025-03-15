# Skyflow Detokenization for Power BI

This repository contains Power BI M functions for detokenizing sensitive data using the Skyflow API. These functions allow you to securely retrieve original values from tokenized data directly within Power BI.

## Contents

- `DetokenizeColumn.pq` - Function to detokenize a single column in a table
- `DetokenizeMultipleColumns.pq` - Function to detokenize multiple columns in a table

## Prerequisites

- Power BI Desktop Pro
- Skyflow Detokenizer Go function up and running either locally, on lambda or on docker
- Basic familiarity with Power BI and M functions

## Setup Instructions

### Step 1: Configure Your Skyflow Detokenization Endpoint

Before using these functions, ensure you have a working Skyflow detokenization service running at one of the following:
- A local server (default: http://localhost:8080/detokenize)
- A deployed Lambda function
- A Docker container

Update the URL in the `DetokenizeColumn` function to match your environment.

### Step 2: Import Functions into Power BI

1. Open Power BI Desktop
2. Go to the "Home" tab and click on "Transform Data" to open Power Query Editor
3. In Power Query Editor, click on "New Source" and select "Blank Query"
4. In the formula bar, delete the default query and paste the content of `DetokenizeColumn.pq`
5. Name this query "DetokenizeColumn"
6. Repeat steps 3-5 for `DetokenizeMultipleColumns.pq`, naming it "DetokenizeMultipleColumns"
7. Right-click on each query and select "Properties"

Optional Steps:

8. Check the "Include in report refresh" option if it's not already checked
9. Also make sure to check "Load enabled" to OFF for both functions, as they are utility functions

### Step 3: Use the Functions in Your Power BI Reports

#### Basic Usage Example

```
let
    // Get your data source
    Source = YourTableName,
    
    // Apply detokenization to selected columns
    Result = DetokenizeMultipleColumns(
        Source,
        {"column1", "column2", "column3"}
    )
in
    Result
```

#### Complete Example

```
let
    // Get your data
    Source = ecom_address,
    
    // Apply detokenization to selected columns
    Result = DetokenizeMultipleColumns(
        Source,
        {"address_line1", "address_line2", "address_landmark", "address_pincode", 
         "contact_first_name", "contact_last_name", "communication_phone"}
    )
in
    Result
```

## How the Functions Work

### DetokenizeColumn

This function:
1. Extracts unique tokens from the specified column
2. Creates a JSON payload for the Skyflow API
3. Makes an API call to your detokenization endpoint
4. Processes the response and adds a new column with detokenized values
5. Returns the table with both the original tokenized column and a new detokenized column (prefixed with "Detokenized_")

### DetokenizeMultipleColumns

This function:
1. Takes a table and a list of column names
2. Applies the `DetokenizeColumn` function sequentially to each column in the list
3. Returns the final table with all detokenized columns

## Customization Options

### Modifying the API Endpoint

In the `DetokenizeColumn` function, locate the following line:

```
"http://localhost:8080/detokenize"
```

Replace it with your Skyflow detokenization endpoint URL.


## Troubleshooting

### Common Issues

1. **API Connection Errors**
   - Verify that your Skyflow detokenization endpoint is running
   - Check network connectivity between Power BI and your endpoint
   - Verify correct URL in the function

2. **Missing or Null Results**
   - Ensure your tokens exist in the Skyflow vault
   - Check token format and ensure it matches what Skyflow expects

3. **Performance Issues**
   - Large datasets may require pagination or batch processing
   - Consider caching detokenization results for frequently used tokens

### Debugging

To debug any issues, you can modify the function to output intermediate results:

1. In Power Query Editor, edit the `DetokenizeColumn` function
2. Add steps that return intermediate values for inspection
3. Use `Error.Record` to capture and display error details

## Security Considerations

- The detokenized data will be visible in Power BI, so ensure appropriate access controls
- Consider implementing row-level security in Power BI for sensitive detokenized data
- Monitor API usage to detect any unusual patterns