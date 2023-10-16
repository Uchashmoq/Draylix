package draylix.server;

import draylix.handler.HttpProxyInitHandler;
import draylix.handler.HttpsProxyInitHandler;
import draylix.handler.Socks5CommandRequestHandler;
import draylix.handler.Socks5InitialRequestHandler;
import io.netty.channel.ChannelHandlerContext;
import io.netty.handler.codec.socksx.v5.Socks5CommandRequestDecoder;
import io.netty.handler.codec.socksx.v5.Socks5InitialRequestDecoder;
import io.netty.handler.codec.socksx.v5.Socks5ServerEncoder;
import io.netty.handler.codec.string.StringDecoder;

import java.nio.charset.StandardCharsets;


public final class Service {
    private Service(){}
    public static final byte DEBUG=0;
    public static final byte HTTPS_PROXY=1;
    public static final byte SOCKS5_PROXY=2;
    public static final byte HTTP_PROXY=3;

    public static final StringDecoder STRING_DECODER=new StringDecoder(StandardCharsets.UTF_8);
    //测试握手以及数据收发
    public static void main(String[] args) {
        byte[] ikey = "H5rruxqFyIf0UdUBhJVrd3Bk8F272KPY".getBytes();
        byte[] iv= "CEwjBM3inZuRqo1B".getBytes();

        Server server = new Server(iv, ikey);
        server.bind("127.0.0.1",9945);
    }

    public static boolean initService(ChannelHandlerContext ctx,byte serviceType) {
        //TODO 初始化服务
        if (serviceType==Service.DEBUG){
            return false;
        }else if(serviceType==Service.HTTPS_PROXY){
            initHttpsProxyService(ctx);
        }else if(serviceType==Service.SOCKS5_PROXY){
            initSocks5ProxyService(ctx);
        }else if(serviceType==Service.HTTP_PROXY){
            initHttpProxyService(ctx);
        }
        else{
            return false;
        }
        return true;
    }

    private static void initHttpProxyService(ChannelHandlerContext ctx) {
        ctx.pipeline()
                .addLast("StringDecoder",STRING_DECODER)
                .addLast(HttpProxyInitHandler.INSTANCE);
    }

    private static void initHttpsProxyService(ChannelHandlerContext ctx){
        ctx.pipeline()
                .addLast(HttpsProxyInitHandler.INSTANCE);
    }

    private static void initSocks5ProxyService(ChannelHandlerContext ctx){
        ctx.pipeline()
                .addLast(Socks5ServerEncoder.DEFAULT)
                .addLast("Socks5InitialRequestDecoder",new Socks5InitialRequestDecoder())
                .addLast(Socks5InitialRequestHandler.INSTANCE)
                .addLast("Socks5CommandRequestDecoder",new Socks5CommandRequestDecoder())
                .addLast(Socks5CommandRequestHandler.INSTANCE);
    }

}
