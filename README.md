# DataAnalyze-Plugins

Some plugins adapted to the DataAnalyze framework

[开发规范](https://x0.pub/view/?view_id=9ca5ace6abff60e83fa9ca10a0b3e828&t=1733465392647)

[框架介绍](https://x0.pub/view/?view_id=33fcb71eeafd765963a840fa191b71e2&t=1733465267113)

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
