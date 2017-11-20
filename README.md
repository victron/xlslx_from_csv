# Wrapper around perl script for statistics parsing

## Objective:
just a simple automation of work. 
`wxp.pl` - script parsing stat files witch located in `STAT` dir, and creates csv files and put them into `CSV` dir.
After those action need to create **xlsx**  files. 

## Requirements

- `wxp.pl` - perl script ()
- configuration files for perl script 
    - hardcoded in code 
        - `my.ref2.safmm.910.2Greport.counters`
        - `my.ref2.safmm.910.3Greport.counters`
- predefined directory structure
```
   ./                                               
    │   com_obj.exe                                 
    │                                               
    ├───CSV                                         
    ├───REPORTS                                     
    ├───SCRIPTS                                     
    │       my.ref2.safmm.910.2Greport.counters     
    │       my.ref2.safmm.910.3Greport.counters     
    │       SGSN_template.xlsx                      
    │       wxp.pl                                  
    │                                               
    └───STAT                                        
                                                    
```
## Auto mode
`auto` subcommand allows to download files automatically and parse when file is available
`auto -h` to get help  

## Known issues and limitations

## Dependencies
[github.com/go-ole/go-ole](https://github.com/go-ole/go-ole)

## Using
`-help | --help` - show long help description

## References
[Go lib excelize](https://github.com/xuri/excelize)


 