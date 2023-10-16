package draylix.server;


import draylix.handler.ActiveHandler;
import draylix.handler.DraylixHandshakeHandler;
import draylix.proctocol.DraylixCodec;
import draylix.proctocol.Verifier;
import draylix.util.VerifierFactory;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.channel.socket.nio.NioSocketChannel;
import io.netty.handler.codec.LengthFieldBasedFrameDecoder;
import io.netty.handler.logging.LogLevel;
import io.netty.handler.logging.LoggingHandler;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import java.net.InetSocketAddress;

public class Server {
    private NioEventLoopGroup boss=new NioEventLoopGroup();
    private NioEventLoopGroup worker = new NioEventLoopGroup();
    private ServerBootstrap serverBootstrap = new ServerBootstrap();
    private Logger logger= LoggerFactory.getLogger(this.getClass());
    public VerifierFactory verifierFactory=new VerifierFactory();

    public Server(byte[] iv,byte[] ikey){
        serverBootstrap
                .group(boss,worker)
                .channel(NioServerSocketChannel.class)
                .childHandler(new ChannelInitializer<NioSocketChannel>() {
                    @Override
                    protected void initChannel(NioSocketChannel ch) throws Exception {
                        ch.pipeline()
                                .addLast(new LengthFieldBasedFrameDecoder(1024*1024,4,4,0,0))
                                .addLast(new ActiveHandler(iv,ikey))
                                .addLast(DraylixCodec.INSTANCE)
                                .addLast(new DraylixHandshakeHandler(verifierFactory.getVerifierImpl()))
                        ;
                    }
                });
    }
    public void bind(String ip,int port){
        serverBootstrap.bind(new InetSocketAddress(ip,port));
        logger.info("listening at {}:{}",ip,port);
    }
    public void bind(String ipPort){
        String[] args = ipPort.split(":");
        if(args.length<2){
            throw new RuntimeException("listenAddress=\"ip:port\"");
        }
        String ip = args[0];
        int port=Integer.parseInt(args[1]);
        bind(ip,port);
    }

}
