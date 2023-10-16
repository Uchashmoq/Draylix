package draylix.handler;

import draylix.server.Visitor;
import io.netty.buffer.ByteBuf;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelInboundHandlerAdapter;
import io.netty.channel.SimpleChannelInboundHandler;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class TransmitHandler extends ChannelInboundHandlerAdapter {
    private Logger logger= LoggerFactory.getLogger(this.getClass());
    private Visitor visitor;

    public  TransmitHandler(Visitor visitor){
        this.visitor=visitor;
    }


    @Override
    public void channelRead(ChannelHandlerContext ctx, Object msg) throws Exception {
        ByteBuf buf=(ByteBuf)msg;
        visitor.send(buf);
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        ctx.close();
        visitor.close();
        logger.error(cause.getMessage());
    }

    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        visitor.close();
        super.channelInactive(ctx);
    }
}
