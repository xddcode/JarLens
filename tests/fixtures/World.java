package example;

public record World(String message) {
    public String describe() {
        return "World: " + message;
    }
}
