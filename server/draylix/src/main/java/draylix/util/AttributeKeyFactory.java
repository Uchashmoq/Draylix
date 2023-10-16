package draylix.util;

import io.netty.util.AttributeKey;

public class AttributeKeyFactory {
    private AttributeKeyFactory(){}
    public static final AttributeKey<String> TOKEN= AttributeKey.valueOf("token");
    public static final AttributeKey<byte[]> DRLX_KEY=AttributeKey.valueOf("draylixKey");
    public static final AttributeKey<byte[]> IV=AttributeKey.valueOf("iv");
}
