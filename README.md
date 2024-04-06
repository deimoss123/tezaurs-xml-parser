# Tēzaura atvērto datu parseris

## Lejupielādēt atvērtos datus

```
wget https://repository.clarin.lv/repository/xmlui/bitstream/handle/20.500.12574/104/tezaurs_2024_2_tei.xml
```

## Rakstīt Postgres datubāzē

`-table` karodziņu var izlaist, ja nevēlies veidot jaunu tabulu
```
go run main.go -f tezaurs_2024_2_tei.xml -pg <url> -table
```

## Rakstīt failā

Faila nosaukums būs `def.txt`
```
go run main.go -f tezaurs_2024_2_tei.xml -w
```

### Meklēt ar fzf pēc definīcijas
```
fzf --delimiter "    " --nth 2 --with-nth 2,3 --literal --no-hscroll < def.txt
```
