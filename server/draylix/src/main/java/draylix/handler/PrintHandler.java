package draylix.handler;

import io.netty.buffer.ByteBuf;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;

public class PrintHandler extends SimpleChannelInboundHandler<ByteBuf> {

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, ByteBuf msg) throws Exception {
        msg.retain();
        for (int i = 0; i <msg.readableBytes() ; i++) {
            System.out.printf("[%d] %d ",i,msg.getByte(i));
            if((i+1)%10==0) System.out.println("\n");
        }
        System.out.println("\n\n\n");
        ctx.fireChannelRead(msg);
    }
}
