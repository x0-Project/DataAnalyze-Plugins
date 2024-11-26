# DataAnalyze-Plugins
Some plugins adapted to the DataAnalyze framework

## How to compile


```
go build -buildmode=plugin -o 01xxx.so Tests/xxx.go
```

## Best Development Practices

### Using Goland File Watcher

![image](https://github.com/user-attachments/assets/c07bbbaf-1085-4f8f-a06a-b3457c74cf59)


## Notice

Please note that the compiled product needs to be prefixed with two digits (01editor.so). This determines the automatic loading order of plugins.

Then please place it in the plugins directory of the same level as the HaEAnalyze-Engine executable file. Restarting the main program will automatically load the plugins according to their order.
