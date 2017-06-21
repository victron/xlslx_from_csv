# Wrapper around perl script for statistics parsing

## Objective:
just a simple automation of work. 
`wxp.pl` - script parsing stat files witch located in `STAT` dir, and creates csv files and put them into `CSV` dir.
After those action need to create **xlsx**  files. 
To avoid errors with **xlsx** files, it's possible to create it using [COM Interface](https://msdn.microsoft.com/en-us/library/windows/desktop/ff485850(v=vs.85).aspx)

## Requirements
- Only Windows platform (in reason **COM**)
- installed MS Office
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
## Known issues and limitations
- `Old format or invalid type library`
    - as workaround need to set regional settings on Windows to `en-US`
    - Too complicated to resolve in code [System.Globalization.CultureInfo](https://github.com/go-ole/go-ole/issues/145)
## Dependencies
[github.com/go-ole/go-ole](https://github.com/go-ole/go-ole)

## Using
`-help | --help` - show long help description

## References
[Go lib for COM interface](https://github.com/go-ole/go-ole)

[wraper around go-ole](https://github.com/aswjh/excel)

[another wraper around go-ole](https://github.com/nivrrex/excel/blob/master/excel.go)

 