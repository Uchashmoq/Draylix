package draylix.proctocol;

import draylix.cipher.AesCodec;
import draylix.util.AttributeKeyFactory;
import io.netty.buffer.ByteBuf;
import io.netty.channel.ChannelHandler;
import io.netty.channel.ChannelHandlerContext;
import io.netty.handler.codec.MessageToMessageCodec;
import io.netty.util.ReferenceCountUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.Arrays;
import java.util.List;
@ChannelHandler.Sharable
public class DraylixCodec extends MessageToMessageCodec<ByteBuf,ByteBuf> {
    private Logger logger = LoggerFactory.getLogger(this.getClass());
    private DraylixCodec(){}
    public static final DraylixCodec INSTANCE=new DraylixCodec();
    public static final byte[] HEAD = "drlx".getBytes();

    @Override
    protected void encode(ChannelHandlerContext ctx, ByteBuf buf, List<Object> out) throws Exception {
        byte[] data = new byte[buf.readableBytes()];
        buf.readBytes(data);
        byte[] key = ctx.channel().attr(AttributeKeyFactory.DRLX_KEY).get();
        byte[] iv = ctx.channel().attr(AttributeKeyFactory.IV).get();
        byte[] cipherText = AesCodec.encrypt(data, key, iv);

        ByteBuf buf1=ctx.alloc().buffer(4+4+cipherText.length);
        buf1.writeBytes(HEAD);
        buf1.writeInt(cipherText.length);
        buf1.writeBytes(cipherText);
        out.add(buf1);
    }

    @Override
    protected void decode(ChannelHandlerContext ctx, ByteBuf buf, List<Object> out) throws Exception {
        if(buf.readableBytes()<4+4){
            logger.error("too short message");
            ctx.channel().close();
            return;
        }

        byte[] head = new byte[4];
        buf.readBytes(head);
        if (!Arrays.equals(HEAD,head)) {
            logger.warn("invalid head {}",head);
            ctx.channel().close();
            return;
        }
        buf.readInt();
        byte[] cipherText = new byte[buf.readableBytes()];
        buf.readBytes(cipherText);
        byte[] key = ctx.channel().attr(AttributeKeyFactory.DRLX_KEY).get();
        byte[] iv = ctx.channel().attr(AttributeKeyFactory.IV).get();

        ByteBuf buf1=ctx.alloc().buffer(cipherText.length);
        buf1.writeBytes(AesCodec.decrypt(cipherText,key,iv));
        out.add(buf1);
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        if (cause instanceof IOException) return;
        logger.error(cause.getMessage());
    }
}
