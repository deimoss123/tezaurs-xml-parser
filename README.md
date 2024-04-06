# Tēzaura atvērto datu parseris

Pagaidām raksta tikai failā, lai to var izmantot ar fzf

Lejupielādēt atvērtos datus
```
wget https://repository.clarin.lv/repository/xmlui/bitstream/handle/20.500.12574/104/tezaurs_2024_2_tei.xml
```

Palaist parsētāju
```
go run main.go -f tezaurs_2024_2_tei.xml
```

Meklēt ar fzf pēc definīcijas
```
fzf --delimiter "    " --nth 2 --with-nth 2,3 --literal --no-hscroll < def.txt
```
