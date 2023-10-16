package draylix.server;

import io.netty.bootstrap.Bootstrap;
import io.netty.buffer.ByteBuf;
import io.netty.channel.*;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.nio.NioSocketChannel;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.net.InetSocketAddress;

public class Visitor {
    private Bootstrap bootstrap=new Bootstrap();
    private NioEventLoopGroup worker = new NioEventLoopGroup();
    private Logger logger= LoggerFactory.getLogger(this.getClass());
    private Channel channel;
    private Channel localChannel;
    String remoteAddr;
    public volatile long traffic = 0;

    public Visitor(Channel localChannel){
        this.localChannel=localChannel;

        bootstrap
                .group(worker)
                .channel(NioSocketChannel.class)
                .handler(new ChannelInitializer<NioSocketChannel>() {
                    @Override
                    protected void initChannel(NioSocketChannel ch) throws Exception {
                        ch.pipeline()
                                .addLast(new ChannelInboundHandlerAdapter(){
                                    @Override
                                    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
                                        logger.debug("visitor disconnected from {}",remoteAddr);
                                        close();
                                    }
                                    @Override
                                    public void channelRead(ChannelHandlerContext ctx, Object msg) throws Exception {
                                        ByteBuf buf=(ByteBuf)msg;
                                        trafficIncrease(buf.readableBytes());
                                        localChannel.writeAndFlush(buf);
                                    }
                                    @Override
                                    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
                                        if (!(cause instanceof IOException)) {
                                            logger.error(cause.getMessage());
                                        }
                                    }
                                });
                    }
                });
    }

    public synchronized void trafficIncrease(long p){
        this.traffic+=p;
    }

    public  void close(){
        if(channel!=null){
            channel.close().addListener(
                    f->worker.shutdownGracefully().addListener(f1->{
                        channel=null;
                        worker=null;
                    })
            );
        }
    }

    public boolean send(ByteBuf buf){
        trafficIncrease(buf.readableBytes());
        if(channel!=null)  {
            channel.writeAndFlush(buf);
            return true;
        }else{
            return false;
        }
    }
    public boolean connect(String host,int port){
        return connect(new InetSocketAddress(host,port));
    }


    public boolean connect(InetSocketAddress address){
        try {
            this.channel=bootstrap.connect(address).sync().channel();
            remoteAddr=channel.remoteAddress().toString();
            return true;
        } catch (InterruptedException e) {
            logger.error("connect error : {}",e.getMessage());
            return false;
        }
    }

}

