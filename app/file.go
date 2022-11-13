ackage main

import (
        "encoding/csv"
        "io"
        "os"
        "fmt"
        "strings"
)

func readFile(path string){
        csv_file, err := os.Open(path)
        if err != nil{
                fmt.Printf("File \"%s\" does not exists, did you mount it?\n",path)
                fmt.Println(err)
                return
        }
        r := csv.NewReader(csv_file)
        i := 0
        f := 0
        for {
                record, err := r.Read()
                if err == io.EOF {
                        fmt.Println("Finished whole file")
                        break
                }
                i ++
                if err != nil {
                        f ++
                        fmt.Println("Error reading line:")
                        fmt.Println(err)
                        continue
                }
                s := server{}
                for i,data := range record {
                        switch i {
                        case 0:
                                if strings.HasPrefix(data, "#") {
                                        //Kommentarzeile
                                        break
                                }else if  len(data) == 0 {
                                        //Leere Zeile zur Struktur
                                        break
                                }
                                s.url = data
                        case 1:
                                s.group = data
                        case 2:
                                s.name = data
                        case 3:
                                s.class = data
                        case 4:
                                s.dns = data
                        case 5:
                                s.expectedIp = data
                        }
                }
                if len(s.url) > 0 {
                        fmt.Printf("Found Server URL: %s with dns %s\n", s.url, s.dns)
                        metrics = append(metrics, s)
                }
        }
        fmt.Printf("Read file %s with %d lines, %d errors, %d servers\n", path, i, f, len(metrics))
}
